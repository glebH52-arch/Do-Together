import { useEffect, useLayoutEffect, useRef, useState, type ReactNode, type RefObject } from 'react';
import { createPortal } from 'react-dom';
import { MoreHorizontal } from 'lucide-react';

export function useModalSurface(ref: RefObject<HTMLElement|null>, onClose:()=>void) {
  const close=useRef(onClose); close.current=onClose;
  useEffect(()=>{
    const previous=document.activeElement instanceof HTMLElement?document.activeElement:null;
    const overflow=document.body.style.overflow; document.body.style.overflow='hidden';
    const items=()=>[...(ref.current?.querySelectorAll<HTMLElement>('button:not(:disabled),input:not(:disabled),textarea:not(:disabled),select:not(:disabled),a[href],summary,[tabindex="0"]')??[])].filter(item=>item.getClientRects().length>0);
    (items()[0]??ref.current)?.focus();
    const key=(event:KeyboardEvent)=>{
      if(event.defaultPrevented)return;
      if(event.key==='Escape'){event.preventDefault();close.current();}
      if(event.key!=='Tab')return;
      const list=items(),first=list[0],last=list.at(-1);
      if(!first){event.preventDefault();ref.current?.focus();return;}
      if(event.shiftKey&&(document.activeElement===first||!ref.current?.contains(document.activeElement))){event.preventDefault();last?.focus();}
      else if(!event.shiftKey&&(document.activeElement===last||!ref.current?.contains(document.activeElement))){event.preventDefault();first.focus();}
    };
    document.addEventListener('keydown',key);
    return()=>{document.body.style.overflow=overflow;document.removeEventListener('keydown',key);if(previous?.isConnected)previous.focus();};
  },[ref]);
}

export function ActionMenu({label,children}:{label:string;children:(close:()=>void)=>ReactNode}){
  const [open,setOpen]=useState(false),[position,setPosition]=useState({left:0,top:0});
  const trigger=useRef<HTMLButtonElement>(null),menu=useRef<HTMLDivElement>(null);
  const close=()=>{setOpen(false);trigger.current?.focus();};
  useLayoutEffect(()=>{
    if(!open)return;
    const update=()=>{const anchor=trigger.current?.getBoundingClientRect(),popup=menu.current?.getBoundingClientRect();if(!anchor||!popup)return;setPosition({left:Math.max(8,Math.min(anchor.right-popup.width,window.innerWidth-popup.width-8)),top:anchor.bottom+popup.height+8<=window.innerHeight?anchor.bottom+4:Math.max(8,anchor.top-popup.height-4)});};
    update();menu.current?.querySelector<HTMLButtonElement>('button:not(:disabled)')?.focus();
    const outside=(event:PointerEvent)=>{if(event.target instanceof Node&&!menu.current?.contains(event.target)&&!trigger.current?.contains(event.target))setOpen(false);};
    const key=(event:KeyboardEvent)=>{
      if(event.key==='Escape'||event.key==='Tab'){
        // The portal is outside the modal's DOM subtree: consume Tab here,
        // dismiss the menu, and return to its trigger before the modal trap runs.
        event.preventDefault();event.stopImmediatePropagation();close();return;
      }
      if(!['ArrowDown','ArrowUp','Home','End'].includes(event.key))return;
      const buttons=[...(menu.current?.querySelectorAll<HTMLButtonElement>('button:not(:disabled)')??[])];
      if(!buttons.length)return;
      event.preventDefault();event.stopImmediatePropagation();
      const current=buttons.indexOf(document.activeElement as HTMLButtonElement);
      const next=event.key==='Home'?0:event.key==='End'?buttons.length-1:(current+(event.key==='ArrowUp'?-1:1)+buttons.length)%buttons.length;
      buttons[next]?.focus();
    };
    document.addEventListener('pointerdown',outside);document.addEventListener('keydown',key,true);window.addEventListener('resize',update);window.addEventListener('scroll',update,true);
    const observer=new ResizeObserver(update);if(menu.current)observer.observe(menu.current);
    return()=>{document.removeEventListener('pointerdown',outside);document.removeEventListener('keydown',key,true);window.removeEventListener('resize',update);window.removeEventListener('scroll',update,true);observer.disconnect();};
  },[open]);
  return <><button type="button" ref={trigger} className="icon-control comment-menu-trigger" aria-label={label} aria-haspopup="menu" aria-expanded={open} onClick={()=>setOpen(value=>!value)}><MoreHorizontal size={18}/></button>{open&&createPortal(<div ref={menu} className="comment-menu-popover" role="menu" aria-label={label} style={{position:'fixed',...position}}>{children(close)}</div>,document.body)}</>;
}
