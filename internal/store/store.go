package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lostf1sh/pomo/internal/timer"
	bolt "go.etcd.io/bbolt"
)

var sessionsBucket = []byte("sessions")

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

func (s *Store) Close() error {
	return s.db.Close()
}
