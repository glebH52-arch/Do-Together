import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { canMoveTask, escapeTarget, friendlyActionError, mergeUniqueByID, moveTaskLocally } from "../src/project-detail-model.ts";

const task = { id:"t1", status:"todo", allowed_transitions:["in_progress"], title:"T" };

test("allowed transition is optimistic and rollback snapshot remains unchanged", () => {
  assert.equal(canMoveTask(task,"in_progress"),true);
  assert.equal(canMoveTask(task,"done"),false);
  const before=[task];
  const optimistic=moveTaskLocally(before,"t1","in_progress");
  assert.equal(optimistic[0].status,"in_progress");
  assert.equal(before[0].status,"todo");
  assert.equal(moveTaskLocally(optimistic,"t1","in_progress").length,1);
});

test("history cursor pages are merged without duplicates", () => {
  assert.deepEqual(mergeUniqueByID([{id:"1"},{id:"2"}],[{id:"2"},{id:"3"}]).map(x=>x.id),["1","2","3"]);
});

test("server failures use calm messages without leaking backend text", () => {
  assert.match(friendlyActionError({status:500,message:"pq: internal"},"fallback"),/временно недоступен/i);
  assert.match(friendlyActionError({status:422},"fallback"),/поля/i);
});

test("Escape closes the topmost private panel", () => {
  assert.equal(escapeTarget({task:true,team:true,activity:true}),"task");
  assert.equal(escapeTarget({task:false,team:true,activity:true}),"team");
  assert.equal(escapeTarget({task:false,team:false,activity:true}),"activity");
});

test("task form retains controlled values on error and guards duplicate submit", async () => {
  const source=await readFile(new URL("../src/ProjectDetail.tsx",import.meta.url),"utf8");
  assert.match(source,/if\(submitting\.current\)return/);
  assert.match(source,/fallback:'saveFailed'/);
  assert.doesNotMatch(source,/prompt\(/);
});

test("filters abort stale requests and invites stay separate from notifications", async () => {
  const [detail,api]=await Promise.all([
    readFile(new URL("../src/ProjectDetail.tsx",import.meta.url),"utf8"),
    readFile(new URL("../src/api.ts",import.meta.url),"utf8"),
  ]);
  assert.match(detail,/loadController\.current\?\.abort\(\)/);
  assert.match(api,/"\/invites\/count"/);
  assert.match(api,/"\/notifications\/unread-count"/);
});
