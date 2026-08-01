CREATE TABLE invites (
    id SERIAL PRIMARY KEY,
    project_id INT NOT NULL,
    inviter_id INT NOT NULL,
    invitee_id INT NOT NULL,
    role TEXT NOT NULL ,
    status TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (inviter_id) REFERENCES users(id),
   FOREIGN KEY (invitee_id) REFERENCES users(id),
   FOREIGN KEY (project_id) REFERENCES projects(id)
);


ALTER TABLE invites
ADD CONSTRAINT invite_role_allowed_check
CHECK (role IN ( 'admin', 'member'));

ALTER TABLE invites
ADD CONSTRAINT invite_status_allowed_check
CHECK (status IN ('pending', 'accepted', 'declined', 'expired'));


ALTER TABLE invites
ADD CONSTRAINT invite_allowed_check
CHECK (inviter_id <> invitee_id);


CREATE UNIQUE INDEX invites_one_pending_idx
ON invites (project_id, invitee_id)
WHERE status = 'pending';