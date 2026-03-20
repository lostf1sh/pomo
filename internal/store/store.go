package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lostf1sh/pomo/internal/timer"
	bolt "go.etcd.io/bbolt"
)

var sessionsBucket = []byte("sessions")
var activeStateBucket = []byte("active_state")
var activeStateKey = []byte("current")

type Store struct {
	db *bolt.DB
}

func New(dbPath string) (*Store, error) {
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(sessionsBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(activeStateBucket)
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("creating buckets: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) SaveSession(sess timer.Session) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		key := sess.StartTime.Format(time.RFC3339Nano)
		data, err := json.Marshal(sess)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), data)
	})
}

func (s *Store) GetSessions(from, to time.Time) ([]timer.Session, error) {
	var sessions []timer.Session

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		c := b.Cursor()

		fromKey := []byte(from.Format(time.RFC3339Nano))
		toKey := []byte(to.Format(time.RFC3339Nano))

		for k, v := c.Seek(fromKey); k != nil; k, v = c.Next() {
			if string(k) > string(toKey) {
				break
			}
			var sess timer.Session
			if err := json.Unmarshal(v, &sess); err != nil {
				return err
			}
			sessions = append(sessions, sess)
		}
		return nil
	})

	return sessions, err
}

func (s *Store) GetSessionsByTask(task string, from, to time.Time) ([]timer.Session, error) {
	all, err := s.GetSessions(from, to)
	if err != nil {
		return nil, err
	}

	var filtered []timer.Session
	for _, sess := range all {
		if sess.Task == task {
			filtered = append(filtered, sess)
		}
	}
	return filtered, nil
}

func (s *Store) GetAllSessions() ([]timer.Session, error) {
	var sessions []timer.Session

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		return b.ForEach(func(k, v []byte) error {
			var sess timer.Session
			if err := json.Unmarshal(v, &sess); err != nil {
				return err
			}
			sessions = append(sessions, sess)
			return nil
		})
	})

	return sessions, err
}

func (s *Store) SaveActiveState(snapshot timer.Snapshot) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(activeStateBucket)
		data, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		return b.Put(activeStateKey, data)
	})
}

func (s *Store) GetActiveState() (*timer.Snapshot, error) {
	var snapshot *timer.Snapshot

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(activeStateBucket)
		data := b.Get(activeStateKey)
		if data == nil {
			return nil
		}

		var loaded timer.Snapshot
		if err := json.Unmarshal(data, &loaded); err != nil {
			return err
		}
		snapshot = &loaded
		return nil
	})

	return snapshot, err
}

func (s *Store) ClearActiveState() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(activeStateBucket)
		return b.Delete(activeStateKey)
	})
}

func (s *Store) Close() error {
	return s.db.Close()
}

// ImportSessions adds sessions whose StartTime key is not already present.
func (s *Store) ImportSessions(sessions []timer.Session) (added, skipped int, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		for _, sess := range sessions {
			key := []byte(sess.StartTime.Format(time.RFC3339Nano))
			if b.Get(key) != nil {
				skipped++
				continue
			}
			data, err := json.Marshal(sess)
			if err != nil {
				return err
			}
			if err := b.Put(key, data); err != nil {
				return err
			}
			added++
		}
		return nil
	})
	return added, skipped, err
}

// ReplaceAllSessions removes all stored sessions and writes the given list.
func (s *Store) ReplaceAllSessions(sessions []timer.Session) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		var keys [][]byte
		if err := b.ForEach(func(k, v []byte) error {
			keys = append(keys, append([]byte(nil), k...))
			return nil
		}); err != nil {
			return err
		}
		for _, k := range keys {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		for _, sess := range sessions {
			key := []byte(sess.StartTime.Format(time.RFC3339Nano))
			data, err := json.Marshal(sess)
			if err != nil {
				return err
			}
			if err := b.Put(key, data); err != nil {
				return err
			}
		}
		return nil
	})
}

// CountSessions returns the number of sessions in the store.
func (s *Store) CountSessions() (int, error) {
	var n int
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		return b.ForEach(func(k, v []byte) error {
			n++
			return nil
		})
	})
	return n, err
}

// DateRange returns the oldest and newest session StartTime values.
// If there are no sessions, both times are zero.
func (s *Store) DateRange() (oldest, newest time.Time, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		c := b.Cursor()
		firstK, _ := c.First()
		if firstK == nil {
			return nil
		}
		lastK, _ := c.Last()
		oldest, err = time.Parse(time.RFC3339Nano, string(firstK))
		if err != nil {
			return err
		}
		newest, err = time.Parse(time.RFC3339Nano, string(lastK))
		return err
	})
	return oldest, newest, err
}
