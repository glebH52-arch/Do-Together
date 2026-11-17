package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	stmtCreateUser                         = "create_user"
	stmtGetUserByID                        = "get_user_by_id"
	stmtGetUserByEmail                     = "get_user_by_email"
	stmtUpdateUserProfile                  = "update_user_profile"
	stmtCreateProject                      = "create_project"
	stmtCreateProjectMember                = "create_project_member"
	stmtGetProjectByID                     = "get_project_by_id"
	stmtGetListProjects                    = "get_list_projects"
	stmtUpdateProject                      = "update_project"
	stmtGetListProjectMembers              = "get_list_project_members"
	stmtUpdateRole                         = "update_role"
	stmtAddProjectMember                   = "add_project_member"
	stmtRemoveProjectMember                = "remove_project_member"
	stmtCreateInvite                       = "create_invite"
	stmtGetListInvites                     = "get_list_invites"
	stmtCountPendingInvites                = "count_pending_invites"
	stmtAcceptInvite                       = "accept_invite"
	stmtDeclineInvite                      = "decline_invite"
	stmtGetInviteForUpdate                 = "get_invite_for_update"
	stmtFindInviteeByEmail                 = "find_invitee_by_email"
	stmtExpireStaleInvites                 = "expire_stale_invites"
	stmtCanInvite                          = "can_invite"
	stmtIsProjectMember                    = "is_project_member"
	stmtGetInvitePresentation              = "get_invite_presentation"
	stmtCreateTask                         = "create_task"
	stmtInsertTaskCreatedHistory           = "insert_task_created_history"
	stmtGetTaskParticipants                = "get_task_participants"
	stmtGetListTask                        = "get_list_task"
	stmtGetTaskByID                        = "get_task_by_id"
	stmtUpdateTask                         = "update_task"
	stmtRemoveTask                         = "remove_task"
	stmtLockTaskForDelete                  = "lock_task_for_delete"
	stmtTaskMemberExists                   = "task_member_exists"
	stmtLockTaskAssignee                   = "lock_task_assignee"
	stmtTaskHistoryInsert                  = "task_history_insert"
	stmtInsertTaskNotifications            = "insert_task_notifications"
	stmtListNotifications                  = "list_notifications"
	stmtCountUnreadNotifications           = "count_unread_notifications"
	stmtMarkNotificationRead               = "mark_notification_read"
	stmtCreateTaskComment                  = "create_task_comment"
	stmtCommentAccess                      = "comment_access"
	stmtListTaskComments                   = "list_task_comments"
	stmtLockOwnComment                     = "lock_own_comment"
	stmtUpdateOwnComment                   = "update_own_comment"
	stmtCommentEditHistory                 = "comment_edit_history"
	stmtDeleteOwnComment                   = "delete_own_comment"
	stmtCommentDeleteHistory               = "comment_delete_history"
	stmtSearchTasks                        = "search_tasks"
	stmtCreateVerificationCode             = "create_verification_code"
	stmtCreateEmailOutbox                  = "create_email_outbox"
	stmtGetLatestVerificationCodeForUpdate = "get_latest_verification_code_for_update"
	stmtUpdateVerificationFailedAttempts   = "update_verification_failed_attempts"
	stmtConsumeVerificationCode            = "consume_verification_code"
	stmtUpdateUserEmailVerification        = "update_user_email_verification"
	stmtEmailClaimBatch                    = "email_claim_batch"
	stmtEmailMarkSent                      = "email_mark_sent"
	stmtEmailMarkFailed                    = "email_mark_failed"
)

var statements = map[string]string{
	stmtCreateUser: `
		INSERT INTO users (username,display_name,email,password_hash,status,created_at)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id
	`,
	stmtGetUserByID: `
		SELECT id,username,display_name,avatar_filename,avatar_updated_at,email,password_hash,status,created_at,updated_at,email_verified_at FROM users
		WHERE id = $1`,
	stmtGetUserByEmail: `SELECT id,username,display_name,avatar_filename,avatar_updated_at,email,password_hash,status,created_at,updated_at,email_verified_at FROM users
		WHERE email = $1`,
	stmtUpdateUserProfile: `UPDATE users
		SET display_name = $2, avatar_filename = $3, avatar_updated_at = $4, updated_at = $5
		WHERE id = $1`,
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
	 u.display_name,
	 u.avatar_filename,
	 u.avatar_updated_at,
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
	AND (actor.role = 'creator' OR (actor.role = 'admin' AND target.role = 'member' AND $1 = 'member'))
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
	AND (pm.role = 'creator' OR (pm.role = 'admin' AND $3 = 'member'))
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
	AND (actor.role = 'creator' OR (actor.role = 'admin' AND target.role = 'member'))
)
	`,
	stmtFindInviteeByEmail: `SELECT id FROM users WHERE email = $1`,
	stmtExpireStaleInvites: `UPDATE invites SET status = 'expired', updated_at = $3
		WHERE project_id = $1 AND invitee_id = $2 AND status = 'pending' AND expires_at <= $3`,
	stmtCanInvite:       `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2 AND role IN ('creator','admin'))`,
	stmtIsProjectMember: `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2)`,
	stmtCreateInvite: `INSERT INTO invites(project_id, inviter_id, invitee_id, role, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
	stmtGetInvitePresentation: `SELECT p.title, u.username
		FROM projects AS p
		JOIN users AS u ON u.id = $2
		WHERE p.id = $1`,
	stmtGetListInvites: `
	SELECT i.id,i.project_id,i.inviter_id,i.invitee_id,i.role,i.status,i.expires_at,i.created_at,i.updated_at,
		p.title,u.username FROM invites AS i
	JOIN projects AS p ON p.id = i.project_id
	JOIN users AS u ON u.id = i.inviter_id
	WHERE i.invitee_id = $1
	  AND i.status = 'pending'
	  AND i.expires_at > NOW()
	ORDER BY i.created_at, i.id
	`,
	stmtCountPendingInvites: `SELECT COUNT(*) FROM invites
		WHERE invitee_id = $1 AND status = 'pending' AND expires_at > NOW()`,
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
	stmtCreateTask: `INSERT INTO tasks(project_id,title,description,created_by,status,created_at,due_at,task_type,priority,assignee_id)
		SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		FROM project_members AS pm
		WHERE pm.project_id=$1 AND pm.user_id=$4
		RETURNING id`,
	stmtInsertTaskCreatedHistory: `INSERT INTO task_history(
			task_id, project_id, actor_id, actor_name, action,
			object_type, object_id, object_title, new_values
		)
		SELECT $1, $2, $3, actor.display_name, 'task.created',
			'task', $1, $4, $5::jsonb
		FROM users AS actor WHERE actor.id = $3
		RETURNING id`,
	stmtGetTaskParticipants: `SELECT creator.display_name, creator.avatar_filename, creator.avatar_updated_at,
		assignee.display_name, assignee.avatar_filename, assignee.avatar_updated_at
		FROM users AS creator
		LEFT JOIN users AS assignee ON assignee.id = $2
		WHERE creator.id = $1`,

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
	RETURNING target.id;
	`,

	stmtGetListTask: `
	SELECT
    t.id, t.project_id, t.title, t.description,
    t.created_by, t.status, t.created_at,
	    t.updated_at, t.due_at,
	    t.task_type, t.priority, t.assignee_id, assignee.display_name,
		assignee.avatar_filename, assignee.avatar_updated_at,
	    creator.display_name, creator.avatar_filename, creator.avatar_updated_at, t.completed_at
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
ORDER BY CASE t.status WHEN 'todo' THEN 0 WHEN 'in_progress' THEN 1 WHEN 'review' THEN 2 ELSE 3 END,
		 t.created_at, t.id
	`,

	stmtGetTaskByID: `
	SELECT
		t.id, t.project_id, t.title, t.description,
		t.created_by, t.status, t.created_at,
		t.updated_at, t.due_at,
		t.task_type, t.priority, t.assignee_id, assignee.display_name,
		assignee.avatar_filename, assignee.avatar_updated_at,
		creator.display_name, creator.avatar_filename, creator.avatar_updated_at, t.completed_at
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
	stmtLockTaskForDelete: `SELECT target.title
		FROM tasks AS target
		JOIN project_members AS member
			ON member.project_id = target.project_id AND member.user_id = $2
		WHERE target.id = $1 AND target.project_id = $3 AND target.created_by = $2
		FOR UPDATE OF target`,

	stmtTaskMemberExists: `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`,
	stmtLockTaskAssignee: `
	SELECT TRUE
	FROM project_members
	WHERE project_id = $1
	  AND user_id = $2
	FOR KEY SHARE
	`,
	stmtTaskHistoryInsert: `INSERT INTO task_history(
			task_id, project_id, actor_id, actor_name, action,
			object_type, object_id, object_title, old_values, new_values
		)
		SELECT $1, $2, $3, actor.display_name, $4,
			'task', $1, $5, $6::jsonb, $7::jsonb
		FROM users AS actor WHERE actor.id = $3
		RETURNING id`,
	stmtInsertTaskNotifications: `INSERT INTO notifications(recipient_id, project_id, task_id, source_history_id, kind, payload)
		SELECT DISTINCT recipient.id, $1::uuid, $2::uuid, $3::uuid, $4::text,
			jsonb_build_object(
				'task_id', $2::uuid,
				'task_title', task.title,
				'project_id', $1::uuid,
				'project_title', project.title,
				'actor_id', $5::uuid,
				'actor_name', actor.display_name
			)
		FROM (VALUES ($6::uuid), ($7::uuid)) AS recipient(id)
		JOIN tasks AS task ON task.id = $2::uuid AND task.project_id = $1::uuid
		JOIN projects AS project ON project.id = task.project_id
		JOIN users AS actor ON actor.id = $5::uuid
		WHERE recipient.id IS NOT NULL AND recipient.id <> $5::uuid
		ON CONFLICT DO NOTHING`,
	stmtListNotifications: `SELECT id, recipient_id, project_id, task_id, invite_id, kind, payload, created_at, read_at
		FROM notifications WHERE recipient_id = $1 AND kind <> 'invitation'
		ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`,
	stmtCountUnreadNotifications: `SELECT COUNT(*) FROM notifications
		WHERE recipient_id = $1 AND read_at IS NULL AND kind <> 'invitation'`,
	stmtMarkNotificationRead: `UPDATE notifications SET read_at = COALESCE(read_at, NOW())
		WHERE id = $1 AND recipient_id = $2 AND kind <> 'invitation'`,
	stmtCreateTaskComment: `WITH inserted AS (
		INSERT INTO task_comments(task_id, author_id, body)
		SELECT t.id, $3, $4 FROM tasks AS t
		JOIN project_members AS pm ON pm.project_id = t.project_id AND pm.user_id = $3
		WHERE t.project_id = $1 AND t.id = $2
		RETURNING id, task_id, author_id, body, created_at, updated_at
	), history AS (
		INSERT INTO task_history(
			task_id, project_id, actor_id, actor_name, action,
			object_type, object_id, object_title, new_values
		)
		SELECT inserted.task_id, task.project_id, inserted.author_id, author.display_name,
			'comment.created', 'comment', inserted.id, task.title,
			jsonb_build_object('comment_id', inserted.id, 'task_id', task.id)
		FROM inserted
		JOIN tasks AS task ON task.id = inserted.task_id
		JOIN users AS author ON author.id = inserted.author_id
		RETURNING task_id
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
	stmtCommentEditHistory: `INSERT INTO task_history(
			task_id, project_id, actor_id, actor_name, action,
			object_type, object_id, object_title, new_values
		)
		SELECT task.id, task.project_id, $2, actor.display_name, 'comment.updated',
			'comment', $3, task.title, jsonb_build_object('comment_id', $3::uuid, 'task_id', task.id)
		FROM tasks AS task JOIN users AS actor ON actor.id = $2
		WHERE task.id = $1`,
	stmtDeleteOwnComment: `DELETE FROM task_comments AS c USING tasks AS t, project_members AS pm
		WHERE c.id = $1 AND c.task_id = $2 AND c.author_id = $4
		AND t.id = c.task_id AND t.project_id = $3
		AND pm.project_id = t.project_id AND pm.user_id = $4
		RETURNING c.body`,
	stmtCommentDeleteHistory: `INSERT INTO task_history(
			task_id, project_id, actor_id, actor_name, action,
			object_type, object_id, object_title, new_values
		)
		SELECT task.id, task.project_id, $2, actor.display_name, 'comment.deleted',
			'comment', $3, task.title, jsonb_build_object('comment_id', $3::uuid, 'task_id', task.id)
		FROM tasks AS task JOIN users AS actor ON actor.id = $2
		WHERE task.id = $1`,
	stmtSearchTasks: `SELECT t.id, t.project_id, t.title, t.description, t.created_by, t.status,
		t.created_at, t.updated_at, t.due_at, t.task_type, t.priority,
		t.assignee_id, assignee.display_name, assignee.avatar_filename, assignee.avatar_updated_at,
		creator.display_name, creator.avatar_filename, creator.avatar_updated_at, t.completed_at
		FROM tasks AS t JOIN users AS creator ON creator.id = t.created_by
		LEFT JOIN users AS assignee ON assignee.id = t.assignee_id
		WHERE t.project_id = $1
		AND EXISTS (SELECT 1 FROM project_members AS pm WHERE pm.project_id = t.project_id AND pm.user_id = $2)
		AND ($3::text = '' OR t.title ILIKE '%' || $3 || '%' OR t.description ILIKE '%' || $3 || '%')
		AND ($4::text = '' OR t.status = $4)
		AND ($5::uuid IS NULL OR t.assignee_id = $5)
		AND ($6::text = '' OR t.priority = $6)
		AND ($7::text = '' OR t.task_type = $7)
		AND ($8::timestamptz IS NULL OR t.due_at < $8)
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
	stmtCreateVerificationCode: `
		INSERT INTO email_verification_codes(user_id,code_hash,expires_at,created_at)
		VALUES($1,$2,$3,$4)
		RETURNING id
	`,
	stmtCreateEmailOutbox: `
		INSERT INTO email_outbox(user_id,recipient,kind,payload,next_attempt_at,created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
	stmtGetLatestVerificationCodeForUpdate: `
		SELECT id,user_id,code_hash,expires_at,consumed_at,failed_attempts,created_at
		FROM email_verification_codes
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1
		FOR UPDATE
	`,
	stmtUpdateVerificationFailedAttempts: `
		UPDATE email_verification_codes
		SET failed_attempts = $2
		WHERE id = $1::uuid
	`,
	stmtConsumeVerificationCode: `
		UPDATE email_verification_codes
		SET consumed_at = $2
		WHERE id = $1::uuid
			AND consumed_at IS NULL
	`,
	stmtUpdateUserEmailVerification: `
		UPDATE users
		SET
			email_verified_at = $2,
			updated_at = $2
		WHERE id = $1
			AND email_verified_at IS NULL
	`,
	stmtEmailClaimBatch: `
		WITH candidates AS (
			SELECT id
			FROM email_outbox
			WHERE sent_at IS NULL
				AND delivery_attempts < 5
				AND next_attempt_at <= $1
				AND (
					locked_at IS NULL
					OR locked_at < $2
				)
			ORDER BY created_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $3
		)
		UPDATE email_outbox AS outbox
		SET
			locked_at = $1,
			locked_by = $4
		FROM candidates
		WHERE outbox.id = candidates.id
		RETURNING
			outbox.id,
			outbox.user_id,
			outbox.recipient,
			outbox.kind,
			outbox.payload,
			outbox.delivery_attempts,
			outbox.next_attempt_at,
			outbox.sent_at,
			outbox.created_at,
			outbox.locked_at,
			outbox.locked_by,
			outbox.last_error;
	`,
	stmtEmailMarkSent: `
		UPDATE email_outbox
		SET
			sent_at = $3,
			locked_at = NULL,
			locked_by = NULL,
			last_error = NULL,
			payload = '{}'::jsonb
		WHERE id = $1::uuid
			AND locked_by = $2
			AND sent_at IS NULL
	`,
	stmtEmailMarkFailed: `
		UPDATE email_outbox
		SET
			delivery_attempts = delivery_attempts + 1,
			next_attempt_at = $3,
			last_error = LEFT($4, 2000),
			locked_at = NULL,
			locked_by = NULL
		WHERE id = $1::uuid
			AND locked_by = $2
			AND sent_at IS NULL
			AND delivery_attempts < 5
	`,
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
