package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	stmtCreateUser              = "create_user"
	stmtGetUserByID             = "get_user_by_id"
	stmtGetUserByEmail          = "get_user_by_email"
	stmtCreateProject           = "create_project"
	stmtCreateProjectMember     = "create_project_member"
	stmtGetProjectByID          = "get_project_by_id"
	stmtGetListProjects         = "get_list_projects"
	stmtUpdateProject           = "update_project"
	stmtGetListProjectMembers   = "get_list_project_members"
	stmtUpdateRole              = "update_role"
	stmtAddProjectMember        = "add_project_member"
	stmtRemoveProjectMember     = "remove_project_member"
	stmtCreateInvite            = "create_invite"
	stmtGetListInvites          = "get_list_invites"
	stmtAcceptInvite            = "accept_invite"
	stmtDeclineInvite           = "decline_invite"
	stmtGetInviteForUpdate      = "get_invite_for_update"
	stmtFindInviteeByEmail      = "find_invitee_by_email"
	stmtExpireStaleInvites      = "expire_stale_invites"
	stmtCanInvite               = "can_invite"
	stmtIsProjectMember         = "is_project_member"
	stmtCreateTask              = "create_task"
	stmtGetListTask             = "get_list_task"
	stmtGetTaskByID             = "get_task_by_id"
	stmtUpdateTask              = "update_task"
	stmtRemoveTask              = "remove_task"
	stmtTaskMemberExists        = "task_member_exists"
	stmtTaskHistoryInsert       = "task_history_insert"
	stmtInsertTaskNotifications = "insert_task_notifications"
	stmtListNotifications       = "list_notifications"
	stmtMarkNotificationRead    = "mark_notification_read"
	stmtLockProjectMember       = "lock_project_member"
	stmtLockTaskPosition        = "lock_task_position"
	stmtListTaskPositions       = "list_task_positions"
	stmtSetTaskPosition         = "set_task_position"
	stmtTouchTask               = "touch_task"
	stmtTaskMoveHistory         = "task_move_history"
	stmtCreateTaskComment       = "create_task_comment"
	stmtCommentAccess           = "comment_access"
	stmtListTaskComments        = "list_task_comments"
	stmtLockOwnComment          = "lock_own_comment"
	stmtUpdateOwnComment        = "update_own_comment"
	stmtCommentEditHistory      = "comment_edit_history"
	stmtDeleteOwnComment        = "delete_own_comment"
	stmtCommentDeleteHistory    = "comment_delete_history"
	stmtSearchTasks             = "search_tasks"
)

var statements = map[string]string{
	stmtCreateUser: `
		INSERT INTO users (username,email,password_hash,status,created_at)
		VALUES($1,$2,$3,$4,$5)
		RETURNING id
	`,
	stmtGetUserByID: `
		SELECT id,username,email,password_hash,status,created_at,updated_at FROM users
		WHERE id = $1`,
	stmtGetUserByEmail: `SELECT id,username,email,password_hash,status,created_at,updated_at FROM users
		WHERE email = $1`,
	stmtCreateProject: `
	INSERT INTO projects(created_by,title,goal,status,created_at)
	VALUES($1,$2,$3,$4,$5)
	RETURNING id
	`,
	stmtCreateProjectMember: `
	INSERT INTO project_members(project_id,user_id,role,joined_at)
	VALUES($1,$2,$3,$4)
	`,
	stmtGetProjectByID: `
		SELECT p.id,p.created_by,p.title,goal,p.status,p.created_at,p.updated_at FROM projects AS p
		JOIN project_members AS pm ON p.id = pm.project_id
		WHERE p.id = $1 AND pm.user_id = $2
	`,
	stmtGetListProjects: `
		SELECT p.id,p.created_by,p.title,goal,p.status,p.created_at,p.updated_at FROM projects AS p
		JOIN project_members AS pm ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND ($2::text IS NULL OR p.status = $2::text)
		ORDER BY p.id
		LIMIT  $3 OFFSET  $4
	`,
	stmtUpdateProject: `
	UPDATE projects AS p
	SET
    title = $1,
    goal = $2,
	status = $3,
    updated_at = $4
	FROM project_members AS pm
	WHERE p.id = $5
	AND pm.project_id = p.id
	AND pm.user_id = $6
	AND pm.role = 'creator'
	`,
	stmtGetListProjectMembers: `
SELECT
    pm.project_id,
    pm.user_id,
	 u.username,
    u.email,
    pm.role,
    pm.joined_at
FROM project_members pm
JOIN users AS u ON u.id = pm.user_id
WHERE pm.project_id = $1
AND EXISTS (
    SELECT 1
    FROM project_members AS access
    WHERE access.project_id = pm.project_id
      AND access.user_id = $2
)
ORDER BY pm.user_id
	`,
	stmtUpdateRole: `
	UPDATE project_members AS target
	SET
	role = $1
	WHERE target.user_id = $2
	AND target.project_id = $3
	AND target.role <> 'creator'
	AND $1 <> 'creator'
	AND EXISTS (
	SELECT 1
	FROM project_members AS actor
	WHERE actor.project_id = target.project_id
	AND actor.user_id = $4
	AND actor.role IN ('creator', 'admin')
	)
 `,
	stmtAddProjectMember: `
	INSERT INTO project_members(project_id,user_id,role,joined_at)
	SELECT $1, $2, $3, $4
	WHERE $3 <> 'creator'
	AND EXISTS (
	SELECT 1
    FROM project_members AS pm
	WHERE pm.project_id = $1
	AND pm.user_id = $5
	AND pm.role IN ('creator', 'admin')
)
	`,
	stmtRemoveProjectMember: `
	DELETE FROM project_members AS target
	WHERE target.user_id = $1
	AND target.project_id = $2
	AND target.role <> 'creator'
	AND EXISTS (
	SELECT 1
    FROM project_members AS actor
	WHERE actor.project_id = target.project_id
	AND actor.user_id = $3
	AND actor.role IN ('creator', 'admin')
)
	`,
	stmtFindInviteeByEmail: `SELECT id FROM users WHERE email = $1`,
	stmtExpireStaleInvites: `UPDATE invites SET status = 'expired', updated_at = $3
		WHERE project_id = $1 AND invitee_id = $2 AND status = 'pending' AND expires_at <= $3`,
	stmtCanInvite:       `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2 AND role IN ('creator','admin'))`,
	stmtIsProjectMember: `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2)`,
	stmtCreateInvite: `
	WITH inserted AS (
	INSERT INTO invites(project_id,inviter_id,invitee_id,role,status,expires_at,created_at)
	SELECT $1,$2,$3,$4,$5,$6,$7
	WHERE $4 IN ('admin', 'member')
	AND EXISTS (
	SELECT 1
    FROM project_members AS pm
	WHERE pm.project_id = $1
	AND pm.user_id = $2
	AND pm.role IN ('creator', 'admin')
)
	 AND NOT EXISTS (
      SELECT 1
      FROM project_members AS target
      WHERE target.project_id = $1
        AND target.user_id = $3
  )
	RETURNING id, project_id, inviter_id, invitee_id
	), notification AS (
	INSERT INTO notifications(recipient_id, project_id, invite_id, kind, payload)
	SELECT invitee_id, project_id, id, 'invitation', jsonb_build_object('invite_id', id)
	FROM inserted ON CONFLICT DO NOTHING RETURNING invite_id
	)
	SELECT inserted.id, project.title, inviter.username
	FROM inserted JOIN notification ON notification.invite_id = inserted.id
	JOIN projects AS project ON project.id = inserted.project_id
	JOIN users AS inviter ON inviter.id = inserted.inviter_id
	`,
	stmtGetListInvites: `
	SELECT i.id,i.project_id,i.inviter_id,i.invitee_id,i.role,i.status,i.expires_at,i.created_at,i.updated_at,
		p.title,u.username FROM invites AS i
	JOIN projects AS p ON p.id = i.project_id
	JOIN users AS u ON u.id = i.inviter_id
	WHERE i.invitee_id = $1
	ORDER BY i.created_at, i.id
	`,
	stmtAcceptInvite: `
	UPDATE invites
SET status = $1,
    updated_at = $2
WHERE id = $3
  AND status = 'pending'
	`,
	stmtDeclineInvite: `
	UPDATE invites
SET status = 'declined',
    updated_at = $3
WHERE id = $1
  AND invitee_id = $2
  AND status = 'pending'
  AND expires_at > $3
	`,

	stmtGetInviteForUpdate: `
	SELECT id, project_id, inviter_id, invitee_id,
       role, status, expires_at, created_at, updated_at
FROM invites
WHERE id = $1
  AND invitee_id = $2
FOR UPDATE
	`,
	stmtCreateTask: `
	WITH allocation AS (
	    UPDATE projects AS p
	    SET next_task_number = next_task_number + 1
	    WHERE p.id = $1
	      AND EXISTS (SELECT 1 FROM project_members AS pm WHERE pm.project_id = p.id AND pm.user_id = $4)
	      AND ($10::int IS NULL OR EXISTS (SELECT 1 FROM project_members AS assigned WHERE assigned.project_id = p.id AND assigned.user_id = $10))
	    RETURNING p.id, p.next_task_number - 1 AS task_number
	), inserted AS (
	    INSERT INTO tasks(project_id, title, description, created_by, status, created_at, due_at, task_number, task_key, position, task_type, priority, assignee_id)
	    SELECT id, $2, $3, $4, $5, $6, $7, task_number,
	           'DOT-' || id || '-' || task_number, task_number * 1024, $8, $9, $10
	    FROM allocation
	    RETURNING id, uuid::text, task_key, task_number, position, title, description, status, created_by, task_type, priority, assignee_id, due_at
	), history AS (
	    INSERT INTO task_history(task_id, actor_id, action, new_values)
	    SELECT id, created_by, 'created', jsonb_build_object('title', title, 'description', description, 'status', status, 'type', task_type, 'priority', priority, 'assignee_id', assignee_id, 'due_at', due_at)
	    FROM inserted RETURNING id, task_id
	), notification AS (
	    INSERT INTO notifications(recipient_id, project_id, task_id, source_history_id, kind, payload)
	    SELECT inserted.assignee_id, $1, inserted.id, history.id, 'assignment_changed', jsonb_build_object('task_id', inserted.id)
	    FROM inserted JOIN history ON history.task_id = inserted.id
	    WHERE inserted.assignee_id IS NOT NULL AND inserted.assignee_id <> inserted.created_by
	    ON CONFLICT DO NOTHING RETURNING id
	)
	SELECT inserted.id, inserted.uuid, inserted.task_key, inserted.task_number, inserted.position, creator.username, assignee.username
	FROM inserted JOIN history ON history.task_id = inserted.id
	JOIN users AS creator ON creator.id = $4
	LEFT JOIN users AS assignee ON assignee.id = inserted.assignee_id
	`,

	stmtUpdateTask: `
	UPDATE tasks AS target
	SET
    title = $1,
    description = $2,
    status = $3,
    updated_at = $4,
	completed_at = $10,
	task_type = $11,
	priority = $12,
	assignee_id = $13,
	due_at = $14
	FROM project_members pm
	WHERE target.id = $5
	AND target.project_id = $6
	AND pm.project_id = target.project_id
	AND pm.user_id = $7
	AND target.status = $8
	AND target.updated_at IS NOT DISTINCT FROM $9
	AND ($13::int IS NULL OR EXISTS (SELECT 1 FROM project_members AS assigned WHERE assigned.project_id = target.project_id AND assigned.user_id = $13))
	RETURNING target.id;
	`,

	stmtGetListTask: `
	SELECT
    t.id, t.project_id, t.title, t.description,
    t.created_by, t.status, t.created_at,
	    t.updated_at, t.due_at, t.uuid::text, t.task_key,
	    t.task_type, t.priority, t.assignee_id, assignee.username,
	    creator.username, t.position, t.completed_at
FROM tasks AS t
	JOIN users AS creator ON creator.id = t.created_by
	LEFT JOIN users AS assignee ON assignee.id = t.assignee_id
WHERE t.project_id = $1
  AND EXISTS (
      SELECT 1
      FROM project_members AS pm
      WHERE pm.project_id = t.project_id
        AND pm.user_id = $2
  )
ORDER BY CASE t.status WHEN 'todo' THEN 0 WHEN 'in_progress' THEN 1 WHEN 'review' THEN 2 ELSE 3 END, t.position, t.id
	`,

	stmtGetTaskByID: `
	SELECT
		t.id, t.project_id, t.title, t.description,
		t.created_by, t.status, t.created_at,
		t.updated_at, t.due_at, t.uuid::text, t.task_key,
		t.task_type, t.priority, t.assignee_id, assignee.username,
		creator.username, t.position, t.completed_at
	FROM tasks AS t
	JOIN users AS creator ON creator.id = t.created_by
	LEFT JOIN users AS assignee ON assignee.id = t.assignee_id
	WHERE t.id = $1
	  AND t.project_id = $2
	  AND EXISTS (
		  SELECT 1
		  FROM project_members AS pm
		  WHERE pm.project_id = t.project_id
		    AND pm.user_id = $3
	  )
`,

	stmtRemoveTask: `
	DELETE FROM tasks AS target
	WHERE id = $1
	AND target.project_id = $2
	AND target.created_by = $3
	AND EXISTS (
	SELECT 1
	FROM project_members AS pm
	WHERE pm.project_id = target.project_id
	  AND pm.user_id = $3
)
	`,
	stmtTaskMemberExists: `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`,
	stmtTaskHistoryInsert: `INSERT INTO task_history(task_id, actor_id, action, old_values, new_values)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb) RETURNING id`,
	stmtInsertTaskNotifications: `INSERT INTO notifications(recipient_id, project_id, task_id, source_history_id, kind, payload)
		SELECT DISTINCT recipient.id, $1::int, $2::int, $3::bigint, $4::text, jsonb_build_object('task_id', $2::int)
		FROM (VALUES ($6::int), ($7::int)) AS recipient(id)
		WHERE recipient.id IS NOT NULL AND recipient.id <> $5::int
		ON CONFLICT DO NOTHING`,
	stmtListNotifications: `SELECT id, recipient_id, project_id, task_id, invite_id, kind, payload, created_at, read_at
		FROM notifications WHERE recipient_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`,
	stmtMarkNotificationRead: `UPDATE notifications SET read_at = COALESCE(read_at, NOW()) WHERE id = $1 AND recipient_id = $2`,
	stmtLockProjectMember: `SELECT p.id FROM projects AS p JOIN project_members AS pm ON pm.project_id = p.id
		WHERE p.id = $1 AND pm.user_id = $2 FOR UPDATE OF p, pm`,
	stmtLockTaskPosition:  `SELECT status, position FROM tasks WHERE id = $1 AND project_id = $2 FOR UPDATE`,
	stmtListTaskPositions: `SELECT id, position FROM tasks WHERE project_id = $1 AND status = $2 ORDER BY position, id FOR UPDATE`,
	stmtSetTaskPosition:   `UPDATE tasks SET position = $1 WHERE id = $2`,
	stmtTouchTask:         `UPDATE tasks SET updated_at = NOW() WHERE id = $1`,
	stmtTaskMoveHistory: `INSERT INTO task_history(task_id, actor_id, action, old_values, new_values)
		VALUES ($1, $2, 'edited', jsonb_build_object('position', $3::bigint), jsonb_build_object('position', $4::bigint))`,
	stmtCreateTaskComment: `WITH inserted AS (
		INSERT INTO task_comments(task_id, author_id, body)
		SELECT t.id, $3, $4 FROM tasks AS t
		JOIN project_members AS pm ON pm.project_id = t.project_id AND pm.user_id = $3
		WHERE t.project_id = $1 AND t.id = $2
		RETURNING id, task_id, author_id, body, created_at, updated_at
	), history AS (
		INSERT INTO task_history(task_id, actor_id, action, new_values)
		SELECT task_id, author_id, 'comment_created', jsonb_build_object('comment_id', id, 'body', body)
		FROM inserted RETURNING task_id
	)
	SELECT i.id, i.task_id, i.author_id, u.username, i.body, i.created_at, i.updated_at
	FROM inserted AS i JOIN history AS h ON h.task_id = i.task_id
	JOIN users AS u ON u.id = i.author_id`,
	stmtCommentAccess: `SELECT EXISTS (SELECT 1 FROM tasks AS t
		JOIN project_members AS pm ON pm.project_id = t.project_id AND pm.user_id = $3
		WHERE t.project_id = $1 AND t.id = $2)`,
	stmtListTaskComments: `SELECT c.id, c.task_id, c.author_id, u.username, c.body, c.created_at, c.updated_at
		FROM task_comments AS c JOIN users AS u ON u.id = c.author_id
		WHERE c.task_id = $1 ORDER BY c.created_at, c.id LIMIT $2 OFFSET $3`,
	stmtLockOwnComment: `SELECT c.body FROM task_comments AS c
		JOIN tasks AS t ON t.id = c.task_id
		JOIN project_members AS pm ON pm.project_id = t.project_id AND pm.user_id = $4
		WHERE c.id = $1 AND c.task_id = $2 AND t.project_id = $3 AND c.author_id = $4 FOR UPDATE OF c`,
	stmtUpdateOwnComment: `UPDATE task_comments SET body = $1, updated_at = NOW() WHERE id = $2`,
	stmtCommentEditHistory: `INSERT INTO task_history(task_id, actor_id, action, old_values, new_values)
		VALUES ($1, $2, 'edited', jsonb_build_object('comment_id', $3::bigint, 'body', $4::text), jsonb_build_object('comment_id', $3::bigint, 'body', $5::text))`,
	stmtDeleteOwnComment: `DELETE FROM task_comments AS c USING tasks AS t, project_members AS pm
		WHERE c.id = $1 AND c.task_id = $2 AND c.author_id = $4
		AND t.id = c.task_id AND t.project_id = $3
		AND pm.project_id = t.project_id AND pm.user_id = $4
		RETURNING c.body`,
	stmtCommentDeleteHistory: `INSERT INTO task_history(task_id, actor_id, action, old_values)
		VALUES ($1, $2, 'comment_deleted', jsonb_build_object('comment_id', $3::bigint, 'body', $4::text))`,
	stmtSearchTasks: `SELECT t.id, t.project_id, t.title, t.description, t.created_by, t.status,
		t.created_at, t.updated_at, t.due_at, t.uuid::text, t.task_key, t.task_type, t.priority,
		t.assignee_id, assignee.username, creator.username, t.position, t.completed_at
		FROM tasks AS t JOIN users AS creator ON creator.id = t.created_by
		LEFT JOIN users AS assignee ON assignee.id = t.assignee_id
		WHERE t.project_id = $1
		AND EXISTS (SELECT 1 FROM project_members AS pm WHERE pm.project_id = t.project_id AND pm.user_id = $2)
		AND ($3::text = '' OR t.title ILIKE '%' || $3 || '%' OR t.description ILIKE '%' || $3 || '%')
		AND ($4::text = '' OR t.status = $4)
		AND ($5::int IS NULL OR t.assignee_id = $5)
		AND ($6::text = '' OR t.priority = $6)
		AND ($7::text = '' OR t.task_type = $7)
		AND ($8::timestamptz IS NULL OR t.due_at <= $8)
		AND ($9::timestamptz IS NULL OR t.due_at >= $9)
		ORDER BY
		CASE WHEN $10::text = 'created_at' AND $11::text = 'asc' THEN t.created_at END ASC NULLS LAST,
		CASE WHEN $10::text = 'created_at' AND $11::text = 'desc' THEN t.created_at END DESC NULLS LAST,
		CASE WHEN $10::text = 'due_at' AND $11::text = 'asc' THEN t.due_at END ASC NULLS LAST,
		CASE WHEN $10::text = 'due_at' AND $11::text = 'desc' THEN t.due_at END DESC NULLS LAST,
		CASE WHEN $10::text = 'title' AND $11::text = 'asc' THEN t.title END ASC NULLS LAST,
		CASE WHEN $10::text = 'title' AND $11::text = 'desc' THEN t.title END DESC NULLS LAST,
		CASE WHEN $10::text = 'priority' AND $11::text = 'asc' THEN CASE t.priority WHEN 'low' THEN 0 WHEN 'medium' THEN 1 WHEN 'high' THEN 2 ELSE 3 END END ASC NULLS LAST,
		CASE WHEN $10::text = 'priority' AND $11::text = 'desc' THEN CASE t.priority WHEN 'low' THEN 0 WHEN 'medium' THEN 1 WHEN 'high' THEN 2 ELSE 3 END END DESC NULLS LAST,
		t.id ASC LIMIT $12 OFFSET $13`,
}

func prepareStatements(ctx context.Context, conn *pgx.Conn) error {
	for name, sql := range statements {
		_, err := conn.Prepare(ctx, name, sql)
		if err != nil {
			return fmt.Errorf("prepare statement %s: %w", name, err)
		}
	}
	return nil
}
