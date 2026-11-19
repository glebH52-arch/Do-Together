import {afterEach,beforeEach,expect,test,vi} from 'vitest';
import {cleanup,render,screen,waitFor,within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {activityDetails,activityTaskID} from '../src/activity-details';
import {ProjectDetail} from '../src/ProjectDetail';
import {api,type Activity,type Task,type User} from '../src/api';

const projectID='11111111-1111-4111-8111-111111111111';
const taskID='22222222-2222-4222-8222-222222222222';
const commentID='33333333-3333-4333-8333-333333333333';
const user:User={id:'u1',username:'owner',display_name:'Owner',avatar_url:null,email:'test@example.test',status:'active'};
const task:Task={id:taskID,title:'User task',description:'Description from user',type:'task',priority:'medium',status:'todo',created_by:user.id,creator_name:'Owner',creator_avatar_url:null,assignee_id:null,assignee_name:null,assignee_avatar_url:null,due_at:null,created_at:'2026-11-20T00:00:00Z',allowed_transitions:['in_progress']};
const event:Activity={id:'event1',type:'task.updated',occurred_at:'2026-11-20T01:00:00Z',actor:{id:user.id,name:'Owner'},object:{id:taskID,type:'task',title:task.title,deleted:false},data:{old:{},new:{changes:{status:{old:'todo',new:'in_progress'},priority:{old:'medium',new:'high'},title:{old:'Old user title',new:'User task'}}}}};
const comment:Activity={...event,id:'event2',type:'comment.created',object:{id:commentID,type:'comment',title:task.title,deleted:false},data:{old:{},new:{task_id:taskID,comment_id:commentID}}};
let events:Activity[];
const json=(data:unknown,status=200)=>new Response(JSON.stringify(data),{status,headers:{'Content-Type':'application/json'}});
beforeEach(()=>{
  events=[event,comment];history.replaceState(null,'','/projects');
  vi.stubGlobal('ResizeObserver',class {observe(){} unobserve(){} disconnect(){}});
  vi.stubGlobal('fetch',vi.fn(async(input:string)=>{
    const path=new URL(input,'http://localhost').pathname;
    if(path.endsWith('/members'))return json([{...user,user_id:user.id,role:'creator',joined_at:task.created_at}]);
    if(path.endsWith('/tasks/search'))return json({items:[task],total:1,limit:20,offset:0});
    if(path.endsWith('/activity'))return json({items:events,next_cursor:null});
    if(path.endsWith('/comments')||path.endsWith('/history'))return json({items:[],total:0,limit:20,offset:0});
    if(path.endsWith(`/tasks/${taskID}`))return json(task);
    if(path.endsWith(`/projects/${projectID}`))return json({id:projectID,title:'Project',goal:'Goal',status:'active',created_by:user.id});
    throw new Error(`Unexpected request ${path}`);
  }));
});
afterEach(()=>{cleanup();vi.restoreAllMocks();vi.unstubAllGlobals();});
const props={projectID,user,accessToken:'token',onBack:vi.fn(),onTask:vi.fn()};

test('activity details expose every nested change in RU and EN without translating user text',()=>{
  expect(activityDetails(event,'en')).toEqual(['Status: To do → In progress','Priority: Medium → High','Title: Old user title → User task']);
  expect(activityDetails(event,'ru')).toEqual(['Статус: К выполнению → В работе','Приоритет: Средний → Высокий','Название: Old user title → User task']);
  expect(activityDetails(comment,'en')).toEqual([]);
  expect(activityDetails({...event,data:{old:{},new:{changes:{description:{changed:true}}}}},'en')).toEqual(['Description changed']);
  expect(activityDetails({...event,data:{old:{},new:{changes:{description:{changed:true}}}}},'ru')).toEqual(['Описание изменено']);
  expect(activityDetails({...event,data:{old:{},new:{changes:{}}}},'en')).toEqual([]);
  expect(activityDetails({...event,data:{old:{},new:{changes:{assignee_id:{old:null,new:commentID}}}}},'en')).toEqual(['Assignee changed']);
});

test('task targets use comment data.task_id, never comment object.id, and skip deleted tasks',()=>{
  expect(activityTaskID(comment)).toBe(taskID);
  expect(activityTaskID({...comment,type:'comment.deleted',object:{...comment.object,deleted:true}})).toBe(taskID);
  expect(activityTaskID({...comment,data:{old:{},new:{comment_id:commentID}}})).toBeNull();
  expect(activityTaskID({...event,object:{...event.object,deleted:true}})).toBeNull();
  expect(activityTaskID(comment,new Set([taskID]))).toBeNull();
});

test.each(['ru','en'] as const)('history shows changes immediately and opens comment task, preserving search (%s)',async(locale)=>{
  const click=userEvent.setup(),onTask=vi.fn(),getTask=vi.spyOn(api,'task');
  render(<ProjectDetail {...props} locale={locale} onTask={onTask}/>);
  await screen.findByRole('heading',{name:task.title});
  const search=screen.getByRole('textbox',{name:locale==='ru'?'Поиск':'Search'});
  await click.type(search,'User');
  const historyTrigger=screen.getByRole('button',{name:locale==='ru'?'История проекта':'Project history'});
  await click.click(historyTrigger);
  const drawer=await screen.findByRole('dialog',{name:locale==='ru'?'История проекта':'Project history'});
  await within(drawer).findByText(locale==='ru'?'Статус: К выполнению → В работе':'Status: To do → In progress');
  expect(drawer.querySelector('details')).toBeNull();
  expect(drawer.textContent).not.toContain(taskID);expect(drawer.textContent).not.toContain(commentID);
  await click.click(within(drawer).getByRole('button',{name:locale==='ru'?/добавил\(а\) комментарий.*Открыть задачу/:/added comment.*Open task/}));
  await screen.findByRole('dialog',{name:locale==='ru'?`Задача ${task.title}`:`Task ${task.title}`});
  expect(screen.queryByRole('dialog',{name:locale==='ru'?'История проекта':'Project history'})).toBeNull();
  expect(getTask).toHaveBeenCalledWith('token',projectID,taskID);
  expect(onTask).toHaveBeenCalledWith(taskID);expect((search as HTMLInputElement).value).toBe('User');
  await click.keyboard('{Escape}');
  await waitFor(()=>expect(document.activeElement).toBe(historyTrigger));
});

test('a stale 404 keeps history open, shows local message, and disables unavailable task navigation',async()=>{
  vi.spyOn(api,'task').mockRejectedValue({status:404});
  const click=userEvent.setup();render(<ProjectDetail {...props} locale="en"/>);
  await screen.findByRole('heading',{name:task.title});await click.click(screen.getByRole('button',{name:'Project history'}));
  const drawer=await screen.findByRole('dialog',{name:'Project history'});
  await click.click(await within(drawer).findByRole('button',{name:/updated task.*Open task/}));
  await within(drawer).findAllByText('This task was deleted or is no longer accessible.');
  expect(screen.getByRole('dialog',{name:'Project history'})).toBe(drawer);
  expect(within(drawer).queryByRole('button',{name:/Open task/})).toBeNull();
});

test('known deleted tasks and their comment events have no navigation',async()=>{
  events=[{...event,type:'task.deleted',object:{...event.object,deleted:true}},comment];
  const getTask=vi.spyOn(api,'task'),click=userEvent.setup();render(<ProjectDetail {...props} locale="en"/>);
  await screen.findByRole('heading',{name:task.title});await click.click(screen.getByRole('button',{name:'Project history'}));
  const drawer=await screen.findByRole('dialog',{name:'Project history'});
  await within(drawer).findByText(/deleted task/);
  expect(within(drawer).queryByRole('button',{name:/Open task/})).toBeNull();expect(getTask).not.toHaveBeenCalled();
});
