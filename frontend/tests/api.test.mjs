import test from 'node:test';
import assert from 'node:assert/strict';
import { acceptInvite, createInvite, declineInvite, listInvites, listProjectMembers, refreshKey, removeMember, request, setSession, updateMemberRole } from '../src/api.ts';

const storage = new Map();
globalThis.sessionStorage = {
  getItem: key => storage.get(key) ?? null,
  setItem: (key, value) => storage.set(key, value),
  removeItem: key => storage.delete(key),
};
const tokens = (access, refresh) => ({ access_token: access, refresh_token: refresh, token_type: 'Bearer', expires_in: 900 });
const json = value => new Response(JSON.stringify(value), { headers: { 'Content-Type': 'application/json' } });

test('parallel 401 responses trigger one refresh and both requests retry', async () => {
  setSession(tokens('old-access', 'old-refresh'));
  let refreshes = 0;
  let retries = 0;
  globalThis.fetch = async (url, options) => {
    if (url === '/api/auth/refresh') {
      refreshes++;
      await new Promise(resolve => setTimeout(resolve, 10));
      return json(tokens('new-access', 'new-refresh'));
    }
    if (options.headers.Authorization === 'Bearer old-access') return new Response(null, { status: 401 });
    assert.equal(options.headers.Authorization, 'Bearer new-access');
    retries++;
    return json([]);
  };
  assert.deepEqual(await Promise.all([
    request('/projects', undefined, 'old-access'),
    request('/projects', undefined, 'old-access'),
  ]), [[], []]);
  assert.equal(refreshes, 1);
  assert.equal(retries, 2);
  assert.equal(sessionStorage.getItem(refreshKey), 'new-refresh');
  setSession(null);
});

test('403 does not refresh; rejected refresh ends the session', async () => {
  setSession(tokens('access', 'refresh'));
  let refreshes = 0;
  globalThis.fetch = async url => {
    if (url === '/api/auth/refresh') {
      refreshes++;
      return new Response(null, { status: 401 });
    }
    return new Response(null, { status: url.endsWith('/members') ? 403 : 401 });
  };
  await assert.rejects(listProjectMembers(3, 'access'), { status: 403 });
  assert.equal(refreshes, 0);
  await assert.rejects(listInvites('access'), { status: 401 });
  assert.equal(refreshes, 1);
  assert.equal(sessionStorage.getItem(refreshKey), null);
});

test('invite and member helpers use existing routes and exact bodies', async () => {
  setSession(tokens('access', 'refresh'));
  const calls = [];
  globalThis.fetch = async (url, options) => {
    calls.push({ url, method: options.method, body: options.body, authorization: options.headers.Authorization });
    if (url === '/api/invites' || (url === '/api/projects/3/members' && options.method === 'GET')) return json([]);
    if (url === '/api/projects/3/invites') return new Response(JSON.stringify({ id: 7 }), { status: 201 });
    return new Response(null, { status: 204 });
  };
  await listInvites('access');
  await acceptInvite(7, 'access');
  await declineInvite(8, 'access');
  await listProjectMembers(3, 'access');
  await createInvite(3, { invitee_id: 9, role: 'member' }, 'access');
  await updateMemberRole(3, 9, 'admin', 'access');
  await removeMember(3, 9, 'access');
  assert.deepEqual(calls.map(({ url, method }) => [url, method]), [
    ['/api/invites', 'GET'],
    ['/api/invites/7/accept', 'POST'],
    ['/api/invites/8/decline', 'POST'],
    ['/api/projects/3/members', 'GET'],
    ['/api/projects/3/invites', 'POST'],
    ['/api/projects/3/members/9', 'PATCH'],
    ['/api/projects/3/members/9', 'DELETE'],
  ]);
  assert.equal(calls[4].body, JSON.stringify({ invitee_id: 9, role: 'member' }));
  assert.equal(calls[5].body, JSON.stringify({ role: 'admin' }));
  assert.ok(calls.every(call => call.authorization === 'Bearer access'));
  setSession(null);
});
