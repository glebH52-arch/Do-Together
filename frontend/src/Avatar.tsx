import { useState } from "react";
import { resolveAvatarURL } from "./api";

type AvatarPerson = { username: string; display_name: string; avatar_url: string | null };
export function Avatar({ user, className = "avatar" }: { user: AvatarPerson; className?: string }) {
  const url = resolveAvatarURL(user.avatar_url);
  const [failedURL, setFailedURL] = useState<string | null>(null);
  const initials = (user.display_name || user.username).trim().split(/\s+/).map(part => part[0]).join("").slice(0, 2).toUpperCase();
  return url && failedURL !== url
    ? <img className={className} src={url} alt="" onError={() => setFailedURL(url)} />
    : <span className={`${className} avatar-fallback`} aria-hidden="true">{initials}</span>;
}
