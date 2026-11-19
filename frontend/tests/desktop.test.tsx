import { afterEach, beforeEach, expect, test, vi } from 'vitest';
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { App } from '../src/App';
import { ProjectDetail, historyText } from '../src/ProjectDetail';
import { Avatar } from '../src/Avatar';
import { api, refreshKey, resolveAvatarURL, setSession, type Task, type User, type History } from '../src/api';

const user: User = {id:'u1',username:'owner',display_name:'Owner',avatar_url:null,email:'owner@example.test',status:'active'};
const projectID='11111111-1111-4111-8111-111111111111';
const task: Task = {id:'22222222-2222-4222-8222-222222222222',project_id:projectID,title:'User title',description:'User description',type:'task',priority:'medium',status:'todo',created_by:'u1',creator_name:'Owner',creator_avatar_url:null,assignee_id:null,assignee_name:null,assignee_avatar_url:null,due_at:null,created_at:'2026-11-20T00:00:00Z',updated_at:'2026-11-20T00:00:00Z',allowed_transitions:['in_progress']} as Task;
const otherTask={...task,id:'33333333-3333-4333-8333-333333333333',title:'Other card'};
const comments=[{id:'c1',author_id:'u1',author_name:'Owner',body:'My comment',created_at:task.created_at},{id:'c2',author_id:'u2',author_name:'Another author',body:'Someone else’s comment',created_at:task.created_at}];
const json=(value:unknown,status=200)=>new Response(JSON.stringify(value),{status,headers:{'Content-Type':'application/json'}});
let serverTasks:Task[];
beforeEach(()=>{
  localStorage.clear();sessionStorage.clear();history.replaceState(null,'','/projects');
  serverTasks=[{...task},{...otherTask}];
  vi.stubGlobal('ResizeObserver',class {observe(){} disconnect(){} unobserve(){}});
  vi.stubGlobal('fetch',vi.fn(async (input:string,init?:RequestInit)=>{
    const url=new URL(input,'http://localhost');
    if(url.pathname.endsWith('/auth/refresh'))return json({access_token:'test',refresh_token:'test',token_type:'Bearer',expires_in:900});
    if(url.pathname.endsWith('/users/me'))return json(user);
    if(url.pathname.endsWith('/members'))return json([{...user,user_id:user.id,role:'creator'}]);
    if(url.pathname.endsWith('/tasks/search'))return json({items:serverTasks,total:serverTasks.length,limit:20,offset:0});
    if(url.pathname.endsWith('/comments'))return json({items:comments,total:2,limit:20,offset:0});
    if(url.pathname.endsWith('/history'))return json({items:[],total:0,limit:50,offset:0});
    if(url.pathname.endsWith('/activity'))return json({items:[],next_cursor:null});
    if(url.pathname.endsWith('/count')||url.pathname.endsWith('/unread-count'))return json({count:0});
    if(url.pathname.endsWith('/invites'))return json([]);
    if(url.pathname.endsWith('/notifications'))return json({items:[],total:0});
    if(url.pathname.endsWith('/projects'))return json([]);
    if(url.pathname.endsWith(`/tasks/${task.id}`)&&init?.method!=='PATCH')return json(task);
    if(url.pathname.endsWith(`/projects/${projectID}`))return json({id:projectID,title:'Project title',goal:'Goal',status:'active',created_by:user.id});
    throw new Error(`Unexpected request: ${init?.method} ${input}`);
  }));
});
afterEach(()=>{cleanup();setSession(null);vi.unstubAllGlobals();});
const props={projectID,user,accessToken:'test',onBack:vi.fn(),onTask:vi.fn()};

test('invites and notifications close after leaving but allow crossing into the panel',async()=>{
  sessionStorage.setItem(refreshKey,'test');render(<App/>);
  const click=userEvent.setup();
  for(const label of ['Приглашения','Уведомления']){
    const trigger=await screen.findByRole('button',{name:label,exact:true});
    await click.click(trigger);
    const popover=screen.getByRole('dialog');
    fireEvent.mouseLeave(trigger);
    await act(async()=>{await new Promise(resolve=>setTimeout(resolve,80))});
    fireEvent.mouseEnter(popover);
    await act(async()=>{await new Promise(resolve=>setTimeout(resolve,260))});
    expect(screen.getByRole('dialog')).toBe(popover);
    fireEvent.mouseLeave(popover);
    expect(screen.getByRole('dialog')).toBe(popover);
    await waitFor(()=>expect(screen.queryByRole('dialog')).toBeNull());
    await click.click(trigger);await click.keyboard('{Escape}');expect(screen.queryByRole('dialog')).toBeNull();
    await click.click(trigger);await click.click(screen.getByRole('heading',{name:'Мои проекты'}));
    expect(screen.queryByRole('dialog')).toBeNull();
  }
});

test('popover focus protects keyboard interaction and switching cancels leave timers',async()=>{
  sessionStorage.setItem(refreshKey,'test');const view=render(<App/>);const click=userEvent.setup();
  const invites=await screen.findByRole('button',{name:'Приглашения',exact:true});
  invites.focus();await click.keyboard('{Enter}');
  const popover=screen.getByRole('dialog');fireEvent.mouseLeave(invites);fireEvent.mouseLeave(popover);
  await act(async()=>{await new Promise(resolve=>setTimeout(resolve,260))});
  expect(screen.getByRole('dialog')).toBe(popover);
  screen.getByRole('button',{name:'Закрыть'}).focus();fireEvent.keyDown(popover,{key:'Tab'});fireEvent.mouseLeave(popover);
  await act(async()=>{await new Promise(resolve=>setTimeout(resolve,260))});
  expect(screen.getByRole('dialog')).toBe(popover);
  await click.keyboard('{Escape}');
  await click.click(invites);fireEvent.mouseLeave(invites);
  await click.click(screen.getByRole('button',{name:'Настройки',exact:true}));
  const settings=screen.getByRole('dialog');fireEvent.mouseLeave(settings);
  await act(async()=>{await new Promise(resolve=>setTimeout(resolve,260))});
  expect(screen.getByRole('dialog')).toBe(settings);
  await click.keyboard('{Escape}');
  await click.click(invites);fireEvent.mouseLeave(invites);view.unmount();
  await act(async()=>{await new Promise(resolve=>setTimeout(resolve,260))});
  expect(screen.queryByRole('dialog')).toBeNull();
});

test('failed move rolls back only its card, denies invalid drop locally, and success clears error',async()=>{
  const update=vi.spyOn(api,'updateTask').mockRejectedValueOnce({status:500});
  render(<ProjectDetail {...props} locale="en"/>);
  await screen.findByRole('heading',{name:task.title});
  const click=userEvent.setup();
  await click.click(screen.getByRole('button',{name:`Change status ${task.title}`}));
  await click.click(screen.getByRole('menuitem',{name:/In progress/i}));
  await waitFor(()=>expect(screen.getByRole('heading',{name:task.title}).closest('article')!.textContent).toContain('returned'));
  expect(update).toHaveBeenCalledTimes(1);
  const card=screen.getByRole('heading',{name:task.title}).closest('article')!;
  expect(card.closest('.kanban-column')!.textContent).toContain('To do');
  expect(screen.getByRole('heading',{name:otherTask.title}).closest('article')!.textContent).not.toContain('returned');
  expect(document.querySelector('main > .error')).toBeNull();
  const done=screen.getByRole('heading',{name:/Done/}).closest('section')!;
  fireEvent.drop(done,{dataTransfer:{getData:()=>task.id}});
  expect(card.textContent).toContain('transition is unavailable');expect(update).toHaveBeenCalledTimes(1);
  update.mockImplementationOnce(async()=>{serverTasks=serverTasks.map(t=>t.id===task.id?{...t,status:'in_progress'}:t);return serverTasks[0];});
  await click.click(screen.getByRole('button',{name:`Change status ${task.title}`}));
  await click.click(screen.getByRole('menuitem',{name:/In progress/i}));
  await waitFor(()=>expect(screen.getByRole('heading',{name:task.title}).closest('.kanban-column')!.textContent).toContain('In progress'));
  expect(screen.getByRole('heading',{name:task.title}).closest('article')!.querySelector('[role="alert"]')!.textContent).toBe('');
  expect(update).toHaveBeenCalledTimes(2);
});

test('locale changes board and open task immediately without discarding form input',async()=>{
  const view=render(<ProjectDetail {...props} locale="ru"/>);const click=userEvent.setup();
  await click.click(await screen.findByRole('button',{name:/User title User description/}));
  await click.click(within(screen.getByRole('dialog')).getByRole('button',{name:'Редактировать'}));
  await click.clear(screen.getByLabelText('Название'));await click.type(screen.getByLabelText('Название'),'Unsaved title');
  view.rerender(<ProjectDetail {...props} locale="en"/>);
  expect(screen.getByRole('heading',{name:/To do/})).toBeTruthy();
  expect(within(screen.getByRole('dialog')).getByRole('heading',{name:'Comments'})).toBeTruthy();
  expect((screen.getByLabelText('Title') as HTMLInputElement).value).toBe('Unsaved title');
  view.rerender(<ProjectDetail {...props} locale="ru"/>);
  expect((screen.getByLabelText('Название') as HTMLInputElement).value).toBe('Unsaved title');
});

test('comment menu is restricted to its author and Escape closes menu before task',async()=>{
  render(<ProjectDetail {...props} locale="en"/>);const click=userEvent.setup();
  await click.click(await screen.findByRole('button',{name:/User title User description/}));
  await screen.findByText('My comment');
  expect(screen.queryByRole('button',{name:'Comment actions Another author'})).toBeNull();
  await click.click(screen.getByRole('button',{name:'Comment actions Owner'}));
  const menu=screen.getByRole('menu');expect(within(menu).getAllByRole('menuitem').map(e=>e.textContent)).toEqual(['Edit','Delete']);
  expect(menu.parentElement).toBe(document.body);
  fireEvent.mouseLeave(menu);expect(screen.getByRole('menu')).toBe(menu);
  await click.keyboard('{Tab}');expect(screen.queryByRole('menu')).toBeNull();
  expect(document.activeElement).toBe(screen.getByRole('button',{name:'Comment actions Owner'}));
  await click.click(screen.getByRole('button',{name:'Comment actions Owner'}));
  await click.keyboard('{Escape}');expect(screen.queryByRole('menu')).toBeNull();expect(screen.getByRole('dialog')).toBeTruthy();
});

test('media resolver supports proxy and configured production origin without credentials',()=>{
  expect(resolveAvatarURL('/avatars/a.jpg?v=1','/api')).toBe('/api/avatars/a.jpg?v=1');
  expect(resolveAvatarURL('/avatars/a.jpg','https://api.example.test')).toBe('https://api.example.test/avatars/a.jpg');
  expect(resolveAvatarURL('https://cdn.example.test/a.jpg','/api')).toBe('https://cdn.example.test/a.jpg');
  expect(resolveAvatarURL('/api/avatars/a.jpg','/api')).toBe('/api/avatars/a.jpg');
  expect(resolveAvatarURL(null)).toBeNull();
});

test('history translates system fields and roles while preserving user text',()=>{
  const entry={action:'member.role_changed',old_values:{role:'member'},new_values:{role:'admin',display_name:'Имя пользователя'}} as unknown as History;
  const en=historyText(entry,'en'),ru=historyText(entry,'ru');
  expect(en).toContain('Moderator');expect(en).toContain('Имя пользователя');expect(en).not.toContain('admin');
  expect(ru).toContain('Модератор');expect(ru).toContain('Имя пользователя');
});

test('avatar resolves relative URL and falls back once, resets on replacement and removal',()=>{
  const view=render(<Avatar user={{...user,avatar_url:'/avatars/a.jpg'}}/>);
  const img=document.querySelector('img')!;expect(img.getAttribute('src')).toBe('/api/avatars/a.jpg');
  fireEvent.error(img);expect(document.querySelector('img')).toBeNull();expect(screen.getByText('O')).toBeTruthy();
  view.rerender(<Avatar user={{...user,avatar_url:'/avatars/b.jpg'}}/>);expect(document.querySelector('img')!.getAttribute('src')).toBe('/api/avatars/b.jpg');
  view.rerender(<Avatar user={user}/>);expect(document.querySelector('img')).toBeNull();
});

test('duplicate drops produce one PATCH and rollback does not undo another successful move',async()=>{
  let rejectFirst:(reason:unknown)=>void=()=>{};
  const update=vi.spyOn(api,'updateTask').mockImplementation((_token,_project,id)=>{
    if(id===task.id)return new Promise((_resolve,reject)=>{rejectFirst=reject;});
    serverTasks=serverTasks.map(t=>t.id===id?{...t,status:'in_progress'}:t);
    return Promise.resolve(serverTasks.find(t=>t.id===id)!);
  });
  render(<ProjectDetail {...props} locale="en"/>);
  await screen.findByRole('heading',{name:task.title});
  const target=screen.getByRole('heading',{name:/In progress/}).closest('section')!;
  act(()=>{fireEvent.drop(target,{dataTransfer:{getData:()=>task.id}});fireEvent.drop(target,{dataTransfer:{getData:()=>task.id}});});
  expect(update).toHaveBeenCalledTimes(1);
  fireEvent.drop(target,{dataTransfer:{getData:()=>otherTask.id}});
  await waitFor(()=>expect(screen.getByRole('heading',{name:otherTask.title}).closest('.kanban-column')!.textContent).toContain('In progress'));
  await act(async()=>{rejectFirst({status:500});});
  await waitFor(()=>expect(screen.getByRole('heading',{name:task.title}).closest('article')!.textContent).toContain('returned'));
  expect(screen.getByRole('heading',{name:otherTask.title}).closest('.kanban-column')!.textContent).toContain('In progress');
  expect(update).toHaveBeenCalledTimes(2);
});

test('drawers retain filters and restore scroll lock on Escape and backdrop dismissal',async()=>{
  document.body.style.overflow='auto';
  render(<ProjectDetail {...props} locale="en"/>);const click=userEvent.setup();
  await screen.findByRole('heading',{name:task.title});
  await click.type(screen.getByRole('textbox',{name:'Search'}),'User');
  await click.click(screen.getByRole('button',{name:'Project history'}));
  expect(screen.getAllByRole('dialog')).toHaveLength(1);expect(document.body.style.overflow).toBe('hidden');
  await click.keyboard('{Escape}');expect(screen.queryByRole('dialog')).toBeNull();expect(document.body.style.overflow).toBe('auto');
  expect((screen.getByRole('textbox',{name:'Search'}) as HTMLInputElement).value).toBe('User');
  await click.click(screen.getByRole('button',{name:'Team'}));
  expect(screen.getAllByRole('dialog')).toHaveLength(1);
  fireEvent.mouseDown(document.querySelector('.team-overlay')!);
  expect(screen.queryByRole('dialog')).toBeNull();expect(document.body.style.overflow).toBe('auto');
});
