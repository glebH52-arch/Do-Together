import { useState } from "react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App, SettingsPanel } from "../src/App";
import { ProjectMembers } from "../src/ProjectMembers";
import { api, ApiError, refreshKey, setSession, type User } from "../src/api";
import { notificationText } from "../src/i18n";

const original: User = { id:"owner", username:"owner", display_name:"Test Owner", avatar_url:null, email:"owner@example.test", status:"active" };
let serverUser: User;
beforeEach(()=>{
  serverUser={...original}; localStorage.clear();sessionStorage.clear();history.replaceState(null,"","/projects");
  vi.spyOn(api,"members").mockImplementation(async()=>[{...serverUser,user_id:serverUser.id,role:"creator",joined_at:"2026-11-20T00:00:00Z"}]);
});
afterEach(()=>{cleanup();setSession(null);vi.unstubAllGlobals();});
function ProfileAndTeam() {
  const [user,setUser]=useState(original);
  return <><SettingsPanel user={user} token="test" locale="en" setLocale={()=>{}} theme="light" setTheme={()=>{}} onUser={setUser} onLogout={()=>{}}/><ProjectMembers projectID="project" user={user} accessToken="test" locale="en"/></>;
}

test("upload, replace and delete avatars refresh profile and open team from API responses",async()=>{
  const upload=vi.spyOn(api,"uploadAvatar").mockImplementation(async(_token,file)=>{
    serverUser={...serverUser,avatar_url:`/avatars/${file.name}?v=${file.lastModified}`};return serverUser;
  });
  const remove=vi.spyOn(api,"deleteAvatar").mockImplementation(async()=>{serverUser={...serverUser,avatar_url:null};return serverUser;});
  const view=render(<ProfileAndTeam/>), click=userEvent.setup();
  await screen.findByText("Test Owner · You");
  expect(view.container.querySelectorAll("img")).toHaveLength(0);
  const input=view.container.querySelector('input[type="file"]') as HTMLInputElement;
  await click.upload(input,new File(["first"],"first.png",{type:"image/png",lastModified:1}));
  await waitFor(()=>expect(Array.from(view.container.querySelectorAll("img")).map(img=>img.getAttribute("src"))).toEqual(["/api/avatars/first.png?v=1","/api/avatars/first.png?v=1"]));
  await click.upload(input,new File(["second"],"second.jpg",{type:"image/jpeg",lastModified:2}));
  await waitFor(()=>expect(Array.from(view.container.querySelectorAll("img")).map(img=>img.getAttribute("src"))).toEqual(["/api/avatars/second.jpg?v=2","/api/avatars/second.jpg?v=2"]));
  await click.click(screen.getByRole("button",{name:"Remove photo"}));
  await waitFor(()=>expect(view.container.querySelectorAll("img")).toHaveLength(0));
  expect(upload).toHaveBeenCalledTimes(2);expect(remove).toHaveBeenCalledTimes(1);expect(api.members).toHaveBeenCalledTimes(4);
  expect(view.container.querySelector(".profile-avatar .avatar-fallback")?.textContent).toBe("TO");
  expect(view.container.querySelector(".member-avatar")?.textContent).toBe("TO");
});

test("failed upload keeps confirmed avatar, failed image shows initials without repeated image mounts",async()=>{
  const failed=vi.spyOn(api,"uploadAvatar").mockRejectedValue(new ApiError("server detail",503));
  const current={...original,avatar_url:"/avatars/current.png"};
  const onUser=vi.fn();
  const props={user:current,token:"test",locale:"en" as const,setLocale:()=>{},theme:"light" as const,setTheme:()=>{},onUser,onLogout:()=>{}};
  const view=render(<SettingsPanel {...props}/>), click=userEvent.setup();
  await click.upload(view.container.querySelector('input[type="file"]') as HTMLInputElement,new File(["new"],"new.png",{type:"image/png"}));
  expect((await screen.findByRole("alert")).textContent).toBe("The service is temporarily unavailable.");
  expect(failed).toHaveBeenCalledTimes(1);expect(onUser).not.toHaveBeenCalled();
  const img=view.container.querySelector("img")!;expect(img.getAttribute("src")).toBe("/api/avatars/current.png");
  fireEvent.error(img);expect(view.container.querySelector("img")).toBeNull();
  view.rerender(<SettingsPanel {...props}/>);expect(view.container.querySelector("img")).toBeNull();
  expect(view.container.querySelector(".avatar-fallback")?.textContent).toBe("TO");
  view.rerender(<SettingsPanel {...props} user={{...current,avatar_url:"/avatars/changed.png"}}/>);
  expect(view.container.querySelector("img")?.getAttribute("src")).toBe("/api/avatars/changed.png");
});

test("App changes RU to EN immediately, persists across remount and returns to Russian",async()=>{
  vi.stubGlobal("fetch",vi.fn(async()=>new Response(JSON.stringify({access_token:"test",refresh_token:"test",token_type:"Bearer",expires_in:900}),{headers:{"Content-Type":"application/json"}})));
  vi.spyOn(api,"profile").mockResolvedValue(original);vi.spyOn(api,"projects").mockResolvedValue([]);
  vi.spyOn(api,"invites").mockResolvedValue([]);vi.spyOn(api,"notifications").mockResolvedValue({items:[],total:0,limit:30,offset:0});
  vi.spyOn(api,"inviteCount").mockResolvedValue({count:0});vi.spyOn(api,"notificationCount").mockResolvedValue({count:0});
  sessionStorage.setItem(refreshKey,"test");let view=render(<App/>);const click=userEvent.setup();
  await click.click(await screen.findByRole("button",{name:"Настройки",exact:true}));
  await click.selectOptions(screen.getByLabelText("Язык"),"en");
  expect(screen.getByRole("heading",{name:"My projects"})).toBeTruthy();expect(localStorage.getItem("dt-locale")).toBe("en");
  view.unmount();view=render(<App/>);
  expect(await screen.findByRole("heading",{name:"My projects"})).toBeTruthy();
  await click.click(screen.getByRole("button",{name:"Settings",exact:true}));await click.selectOptions(screen.getByLabelText("Language"),"ru");
  expect(screen.getByRole("heading",{name:"Мои проекты"})).toBeTruthy();expect(localStorage.getItem("dt-locale")).toBe("ru");
});

test("backend notification enums render localized actions while preserving user task titles",()=>{
  expect(notificationText("assignment_changed",{task_title:"Пользовательский заголовок"},"en")).toBe("Task assignee changed: Пользовательский заголовок");
  expect(notificationText("status_changed",{},"en")).toBe("Task status changed");
  expect(notificationText("due_at_changed",{},"en")).toBe("Task due date changed");
  expect(notificationText("invitation",{},"ru")).toBe("Приглашение в проект");
});
