package board

import (
	"context"
	"time"
)

const accessTTL = time.Minute

type knownAccess struct {
	canWrite bool
	checked  time.Time
}

func (s *Service) CanWrite(ctx context.Context) (bool, error) {
	editor, ok := s.source.(TaskEditor)
	if !ok {
		return false, nil
	}
	token := tokenDigest(ctx)
	s.mu.Lock()
	known, found := s.access[token]
	s.mu.Unlock()
	if found && s.now().Sub(known.checked) < accessTTL {
		return known.canWrite, nil
	}
	canWrite, err := editor.CanWrite(ctx)
	if err != nil {
		return false, err
	}
	s.mu.Lock()
	for key, entry := range s.access {
		if s.now().Sub(entry.checked) >= accessTTL {
			delete(s.access, key)
		}
	}
	s.access[token] = knownAccess{canWrite: canWrite, checked: s.now()}
	s.mu.Unlock()
	return canWrite, nil
}

func (s *Service) requireEditor(ctx context.Context) (TaskEditor, error) {
	canWrite, err := s.CanWrite(ctx)
	if err != nil {
		return nil, err
	}
	if !canWrite {
		return nil, ErrForbidden
	}
	return s.source.(TaskEditor), nil
}
