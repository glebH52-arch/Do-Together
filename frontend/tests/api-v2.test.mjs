import test from "node:test";
import assert from "node:assert/strict";
import { api, refreshKey, request, setSession } from "../src/api.ts";
const storage = new Map();
globalThis.sessionStorage = {
  getItem: (k) => storage.get(k) ?? null,
  setItem: (k, v) => storage.set(k, v),
  removeItem: (k) => storage.delete(k),
};
const tokens = (a, r) => ({
  access_token: a,
  refresh_token: r,
  token_type: "Bearer",
  expires_in: 900,
});
const json = (v) =>
  new Response(JSON.stringify(v), {
    headers: { "Content-Type": "application/json" },
  });
test("parallel 401 performs one refresh and keeps UUID paths unchanged", async () => {
  setSession(tokens("old", "refresh"));
  let refreshes = 0;
  const calls = [];
  globalThis.fetch = async (url, o) => {
    calls.push([url, o.method, o.headers.Authorization]);
    if (url === "/api/auth/refresh") {
      refreshes++;
      await new Promise((r) => setTimeout(r, 5));
      return json(tokens("new", "next"));
    }
    if (o.headers.Authorization === "Bearer old")
      return new Response(null, { status: 401 });
    return json({ items: [], total: 0, limit: 20, offset: 0 });
  };
  const id = "018f47d2-a9f1-7a2b-8c3d-123456789abc";
  await Promise.all([api.tasks("old", id), api.tasks("old", id)]);
  assert.equal(refreshes, 1);
  assert.ok(calls.some((v) => v[0] === `/api/projects/${id}/tasks/search`));
  assert.equal(sessionStorage.getItem(refreshKey), "next");
  setSession(null);
});
test("nullable UUID and deadline are serialized as null, omitted values stay omitted", async () => {
  setSession(tokens("access", "refresh"));
  let body;
  globalThis.fetch = async (_url, o) => {
    body = JSON.parse(o.body);
    return json({ id: "x" });
  };
  await api.updateTask("access", "project", "task", {
    assignee_id: null,
    due_at: null,
    title: "T",
  });
  assert.deepEqual(body, { assignee_id: null, due_at: null, title: "T" });
  setSession(null);
});
test("query encodes UUID filters and numeric pagination", async () => {
  setSession(tokens("access", "refresh"));
  let url;
  globalThis.fetch = async (u) => {
    url = u;
    return json({ items: [], total: 0, limit: 20, offset: 20 });
  };
  await api.tasks("access", "p", {
    assignee_id: "user-uuid",
    limit: 20,
    offset: 20,
    q: "две строки",
  });
  assert.match(url, /assignee_id=user-uuid/);
  assert.match(url, /offset=20/);
  assert.match(url, /q=/);
  setSession(null);
});
test("invalid refresh clears the persisted session", async () => {
  setSession(tokens("old", "bad"));
  globalThis.fetch = async (u) => new Response(null, { status: 401 });
  await assert.rejects(request("/projects", { token: "old" }), { status: 401 });
  assert.equal(sessionStorage.getItem(refreshKey), null);
});
