import type { Activity, ProjectMember, TaskPriority, TaskType } from './api';
import { activityCopy, activityDetailCopy, boardCopy, fmt, roleLabel, statusLabel, type Locale } from './i18n';

const uuidPattern=/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
export function activityTaskID(event:Activity,deletedTaskIDs:ReadonlySet<string>=new Set()):string|null {
  let id:unknown=null;
  if(event.object.type==='task') {
    if(event.object.deleted||event.type==='task.deleted')return null;
    id=event.object.id;
  } else if(event.object.type==='comment') {
    // object.id is the comment ID, including for deleted comments.
    id=event.data.new?.task_id??event.data.old?.task_id;
  }
  return typeof id==='string'&&uuidPattern.test(id)&&!deletedTaskIDs.has(id)?id:null;
}

/** Format only user-facing fields. IDs locate entities; they are never printed. */
export function activityDetails(event:Activity,locale:Locale,members:ProjectMember[]=[]):string[] {
  const copy=boardCopy[locale],detail=activityDetailCopy[locale],activity=activityCopy[locale];
  const before=event.data.old??{},after=event.data.new??{};
  const visible=new Set(['title','description','goal','status','type','priority','role','display_name','username','invitee_name','assignee_name','assignee_id','due_at','body']);
  const memberName=(id:unknown)=>members.find(member=>member.user_id===id)?.display_name||members.find(member=>member.user_id===id)?.username;
  const text=(key:string,value:unknown):string|null=>{
    if(value==null||value==='')return key==='assignee_id'?copy.unassigned:'—';
    if(key==='assignee_id')return memberName(value)??null;
    if(typeof value!=='string'&&typeof value!=='number'&&typeof value!=='boolean')return null;
    if(key==='role')return ['creator','admin','member'].includes(String(value))?roleLabel(String(value),locale):null;
    if(key==='status')return (activity.statuses as Record<string,string>)[String(value)]??(['todo','in_progress','review','done'].includes(String(value))?statusLabel(String(value),locale):null);
    if(key==='priority')return copy.priorities[value as TaskPriority]??null;
    if(key==='type')return copy.types[value as TaskType]??null;
    if(key==='due_at')return Number.isNaN(new Date(String(value)).getTime())?null:fmt(String(value),locale);
    return String(value);
  };
  const format=(key:string,oldValue:unknown,newValue:unknown,hasOld:boolean,changedOnly=false):string|null=>{
    if(!visible.has(key))return null;
    if(changedOnly)return key==='description'?detail.descriptionChanged:`${(activity.fields as Record<string,string>)[key]}: ${activity.changed}`;
    if(hasOld&&JSON.stringify(oldValue)===JSON.stringify(newValue))return null;
    const oldText=text(key,oldValue),newText=text(key,newValue);
    if(key==='assignee_id'&&(oldText===null||newText===null))return detail.assigneeChanged;
    if(newText===null||hasOld&&oldText===null)return null;
    const label=(activity.fields as Record<string,string>)[key];
    return `${label}: ${hasOld?`${oldText} → `:''}${newText}`;
  };
  const nested=after.changes;
  if(nested&&typeof nested==='object'&&!Array.isArray(nested))return Object.entries(nested).flatMap(([key,value])=>{
    if(!value||typeof value!=='object')return [];
    const change=value as {old?:unknown;new?:unknown;changed?:boolean};
    const line=format(key,change.old,change.new,'old' in change,change.changed===true);
    return line?[line]:[];
  });
  return [...new Set([...Object.keys(before),...Object.keys(after)])].flatMap(key=>{
    // Names are already in the event title; IDs only help resolve task/member links.
    if(key==='title'&&after.title===event.object.title&&!('title' in before))return [];
    const line=format(key,before[key],after[key],key in before);
    return line?[line]:[];
  });
}
