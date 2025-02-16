package models

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"git.kirsle.net/apps/barertc/pkg/config"
	"git.kirsle.net/apps/barertc/pkg/log"
	"git.kirsle.net/apps/barertc/pkg/messages"
)

type DirectMessage struct {
	MessageID int64
	ChannelID string
	Username  string
	Message   string
	Timestamp int64
}

const DirectMessagePerPage = 20

func (dm DirectMessage) CreateTable() error {
	if DB == nil {
		return ErrNotInitialized
	}

	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS direct_messages (
			message_id INTEGER PRIMARY KEY,
			channel_id TEXT,
			username TEXT,
			message TEXT,
			timestamp INTEGER
		);

		CREATE INDEX IF NOT EXISTS idx_direct_messages_channel_id ON direct_messages(channel_id);
		CREATE INDEX IF NOT EXISTS idx_direct_messages_timestamp ON direct_messages(timestamp);
	`)
	if err != nil {
		return err
	}

	// Delete old messages past the retention period.
	if days := config.Current.DirectMessageHistory.RetentionDays; days > 0 {
		cutoff := time.Now().Add(time.Duration(-days) * 24 * time.Hour)
		log.Info("Deleting old DM history past %d days (cutoff: %s)", days, cutoff.Format(time.RFC3339))
		_, err := DB.Exec(
			"DELETE FROM direct_messages WHERE timestamp < ?",
			cutoff.Unix(),
		)
		if err != nil {
			log.Error("Error removing old DMs: %s", err)
		}
	}

	return nil
}

// LogMessage adds a message to the DM history between two users.
func (dm DirectMessage) LogMessage(fromUsername, toUsername string, msg messages.Message) error {
	if DB == nil {
		return ErrNotInitialized
	}

	if msg.MessageID == 0 {
		return errors.New("message did not have a MessageID")
	}

	var (
		channelID = CreateChannelID(fromUsername, toUsername)
		timestamp = time.Now().Unix()
	)

	_, err := DB.Exec(`
		INSERT INTO direct_messages (message_id, channel_id, username, message, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, msg.MessageID, channelID, fromUsername, msg.Message, timestamp)

	return err
}

// ClearMessages clears all stored DMs that the username as a participant in.
func (dm DirectMessage) ClearMessages(username string) (int, error) {
	if DB == nil {
		return 0, ErrNotInitialized
	}

	var placeholders = []interface{}{
		fmt.Sprintf("@%s:%%", username), // `@alice:%`
		fmt.Sprintf("%%:@%s", username), // `%:@alice`
		username,
	}

	// Count all the messages we'll delete.
	var (
		count int
		row   = DB.QueryRow(`
			SELECT COUNT(message_id)
			FROM direct_messages
			WHERE (channel_id LIKE ? OR channel_id LIKE ?)
			OR username = ?
		`, placeholders...)
	)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	// Delete them all.
	_, err := DB.Exec(`
		DELETE FROM direct_messages
		WHERE (channel_id LIKE ? OR channel_id LIKE ?)
		OR username = ?
	`, placeholders...)

	return count, err
}

// TakebackMessage removes a message by its MID from the DM history.
//
// Because the MessageID may have been from a previous chat session, the server can't immediately
// verify the current user had permission to take it back. This function instead will check whether
// a DM history exists sent by this username for that messageID, and if so, returns a
// boolean true that the username/messageID matched which will satisfy the permission check
// in the OnTakeback handler.
func (dm DirectMessage) TakebackMessage(username string, messageID int64, isAdmin bool) (bool, error) {
	if DB == nil {
		return false, ErrNotInitialized
	}

	// Does this messageID exist as sent by the user?
	if !isAdmin {
		var (
			row = DB.QueryRow(
				"SELECT message_id FROM direct_messages WHERE username = ? AND message_id = ?",
				username, messageID,
			)
			foundMsgID int64
			err        = row.Scan(&foundMsgID)
		)
		if err != nil {
			return false, errors.New("no such message ID found as owned by that user")
		}
	}

	// Delete it.
	_, err := DB.Exec(
		"DELETE FROM direct_messages WHERE message_id = ?",
		messageID,
	)

	// Return that it was successfully validated and deleted.
	return err == nil, err
}

// PaginateDirectMessages returns a page of messages, the count of remaining, and an error.
func PaginateDirectMessages(fromUsername, toUsername string, beforeID int64) ([]messages.Message, int, error) {
	if DB == nil {
		return nil, 0, ErrNotInitialized
	}

	var (
		result    = []messages.Message{}
		channelID = CreateChannelID(fromUsername, toUsername)

		// Compute the remaining messages after finding the final messageID this page.
		lastMessageID int64
		remaining     int
	)

	if beforeID == 0 {
		beforeID = math.MaxInt64
	}

	rows, err := DB.Query(`
		SELECT message_id, username, message, timestamp
		FROM direct_messages
		WHERE channel_id = ?
		AND message_id < ?
		ORDER BY message_id DESC
		LIMIT ?
	`, channelID, beforeID, DirectMessagePerPage)
	if err != nil {
		return nil, 0, err
	}

	for rows.Next() {
		var row DirectMessage
		if err := rows.Scan(
			&row.MessageID,
			&row.Username,
			&row.Message,
			&row.Timestamp,
		); err != nil {
			return nil, 0, err
		}

		msg := messages.Message{
			MessageID: row.MessageID,
			Username:  row.Username,
			Message:   row.Message,
			Timestamp: time.Unix(row.Timestamp, 0).Format(time.RFC3339),
		}
		result = append(result, msg)
		lastMessageID = msg.MessageID
	}

	// Get a count of the remaining messages.
	row := DB.QueryRow(`
		SELECT COUNT(message_id)
		FROM direct_messages
		WHERE channel_id = ?
		AND message_id < ?
	`, channelID, lastMessageID)
	if err := row.Scan(&remaining); err != nil {
		return nil, 0, err
	}

	return result, remaining, nil
}

// CreateChannelID returns a deterministic channel ID for a direct message conversation.
//
// The usernames (passed in any order) are sorted alphabetically and composed into the channel ID.
func CreateChannelID(fromUsername, toUsername string) string {
	var parts = []string{fromUsername, toUsername}
	sort.Strings(parts)
	return fmt.Sprintf(
		"@%s:@%s",
		parts[0],
		parts[1],
	)
}
