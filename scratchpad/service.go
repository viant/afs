package scratchpad

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/viant/afs"
	afsurl "github.com/viant/afs/url"
)

// ErrNoteNotFound identifies a missing scratchpad note.
var ErrNoteNotFound = errors.New("scratchpad note not found")

// NoteNotFoundError describes the logical note and resolved storage location
// that could not be found.
type NoteNotFoundError struct {
	Key   string
	URI   string
	Cause error
}

func (e *NoteNotFoundError) Error() string {
	message := ErrNoteNotFound.Error()
	if e == nil {
		return message
	}
	details := make([]string, 0, 2)
	if e.Key != "" {
		details = append(details, fmt.Sprintf("key=%q", e.Key))
	}
	if e.URI != "" {
		details = append(details, fmt.Sprintf("uri=%q", e.URI))
	}
	if len(details) > 0 {
		message += " (" + strings.Join(details, ", ") + ")"
	}
	if e.Cause != nil {
		message += ": " + e.Cause.Error()
	}
	return message
}

func (e *NoteNotFoundError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *NoteNotFoundError) Is(target error) bool {
	return target == ErrNoteNotFound
}

func New(opts ...Option) *Service {
	cfg := Config{
		FS:              afs.New(),
		RootURITemplate: DefaultRootURITemplate,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.FS == nil {
		cfg.FS = afs.New()
	}
	if strings.TrimSpace(cfg.RootURITemplate) == "" {
		cfg.RootURITemplate = DefaultRootURITemplate
	}
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	s := &Service{
		fs:             cfg.FS,
		rootTemplate:   strings.TrimSpace(cfg.RootURITemplate),
		userID:         strings.TrimSpace(cfg.UserID),
		userIDProvider: cfg.UserIDProvider,
		basePath:       strings.TrimSpace(cfg.BasePath),
		macros:         map[string]string{},
		now:            cfg.Now,
	}
	for k, v := range cfg.Macros {
		s.macros[k] = v
	}
	if len(cfg.AllowedTargetSchemes) > 0 {
		s.allowedTargetSchemes = map[string]bool{}
		for _, scheme := range normalizeSchemes(cfg.AllowedTargetSchemes) {
			s.allowedTargetSchemes[scheme] = true
		}
	}
	return s
}

func (s *Service) Memorize(ctx context.Context, input *MemorizeInput) (*MemorizeOutput, error) {
	output := &MemorizeOutput{}
	key := strings.TrimSpace(input.Key)
	description := strings.TrimSpace(input.Description)
	body := strings.TrimSpace(input.Body)
	if key == "" {
		return output, fmt.Errorf("key is required")
	}
	if description == "" {
		return output, fmt.Errorf("description is required")
	}
	if body == "" {
		return output, fmt.Errorf("body is required")
	}
	root, userID, err := s.ResolveRootURIContext(ctx)
	if err != nil {
		return output, err
	}
	if err = s.ensureRoot(ctx, root); err != nil {
		return output, fmt.Errorf("scratchpad storage unavailable: %w", err)
	}
	updatedAt := s.now().UTC().Format(time.RFC3339Nano)
	note := Note{
		Key:         key,
		Description: description,
		Body:        input.Body,
		UserID:      userID,
		UpdatedAt:   updatedAt,
	}
	if err = s.writeNote(ctx, root, key, note); err != nil {
		return output, err
	}
	output.Key = key
	output.Description = description
	output.UpdatedAt = updatedAt
	return output, nil
}

func (s *Service) Append(ctx context.Context, input *AppendInput) (*AppendOutput, error) {
	output := &AppendOutput{}
	key := strings.TrimSpace(input.Key)
	body := strings.TrimSpace(input.Body)
	if key == "" {
		return output, fmt.Errorf("key is required")
	}
	if body == "" {
		return output, fmt.Errorf("body is required")
	}
	root, userID, err := s.ResolveRootURIContext(ctx)
	if err != nil {
		return output, err
	}
	if err = s.ensureRoot(ctx, root); err != nil {
		return output, fmt.Errorf("scratchpad storage unavailable: %w", err)
	}
	target := NoteURL(root, key)
	exists, err := s.fs.Exists(ctx, target)
	if err != nil {
		return output, fmt.Errorf("read scratchpad note failed: %w", err)
	}
	description := strings.TrimSpace(input.Description)
	updatedAt := s.now().UTC().Format(time.RFC3339Nano)
	var note Note
	if exists {
		existing, err := s.readNoteURL(ctx, key, target)
		if err != nil {
			return output, err
		}
		if existing.Key != key {
			return output, fmt.Errorf("scratchpad note identity mismatch for key %q", key)
		}
		if err := validateNoteUser(existing, userID); err != nil {
			return output, err
		}
		note = *existing
		note.Body = AppendNoteBody(note.Body, input.Body)
		if description != "" {
			note.Description = description
		}
		if strings.TrimSpace(note.Description) == "" {
			note.Description = key
		}
	} else {
		output.Created = true
		if description == "" {
			description = key
		}
		note = Note{
			Key:         key,
			Description: description,
			Body:        input.Body,
			UserID:      userID,
		}
	}
	note.UserID = userID
	note.UpdatedAt = updatedAt
	if err = s.writeNote(ctx, root, key, note); err != nil {
		return output, err
	}
	output.Key = note.Key
	output.Description = note.Description
	output.UpdatedAt = note.UpdatedAt
	return output, nil
}

func (s *Service) List(ctx context.Context) (*ListOutput, error) {
	output := &ListOutput{}
	root, userID, err := s.ResolveRootURIContext(ctx)
	if err != nil {
		return output, err
	}
	if ok, err := s.fs.Exists(ctx, root); err != nil {
		return output, fmt.Errorf("list scratchpad notes failed: %w", err)
	} else if !ok {
		output.Entries = []Entry{}
		return output, nil
	}
	objects, err := s.fs.List(ctx, root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			output.Entries = []Entry{}
			return output, nil
		}
		return output, fmt.Errorf("list scratchpad notes failed: %w", err)
	}
	var entries []Entry
	for _, object := range objects {
		if object == nil || object.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(object.Name()), ".json") {
			continue
		}
		note, err := s.readNoteURL(ctx, "", object.URL())
		if err != nil {
			return output, err
		}
		if err := validateNoteUser(note, userID); err != nil {
			return output, err
		}
		entries = append(entries, Entry{
			Key:         note.Key,
			Description: note.Description,
			UpdatedAt:   note.UpdatedAt,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	output.Entries = entries
	return output, nil
}

func (s *Service) Fetch(ctx context.Context, key string) (*FetchOutput, error) {
	output := &FetchOutput{}
	key = strings.TrimSpace(key)
	if key == "" {
		return output, fmt.Errorf("key is required")
	}
	note, err := s.fetchNote(ctx, key)
	if err != nil {
		return output, err
	}
	output.Key = note.Key
	output.Description = note.Description
	output.Body = note.Body
	output.UpdatedAt = note.UpdatedAt
	return output, nil
}

func (s *Service) Delete(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("key is required")
	}
	root, userID, err := s.ResolveRootURIContext(ctx)
	if err != nil {
		return err
	}
	target := NoteURL(root, key)
	exists, err := s.fs.Exists(ctx, target)
	if err != nil {
		return fmt.Errorf("read scratchpad note failed: %w", err)
	}
	if exists {
		note, err := s.readNoteURL(ctx, key, target)
		if err != nil {
			return err
		}
		if note.Key != key {
			return fmt.Errorf("scratchpad note identity mismatch for key %q", key)
		}
		if err := validateNoteUser(note, userID); err != nil {
			return err
		}
	}
	if err := s.fs.Delete(ctx, target); err != nil {
		return fmt.Errorf("delete scratchpad note failed: %w", err)
	}
	return nil
}

func (s *Service) FetchArtifact(ctx context.Context, artifactID string) (*Artifact, error) {
	artifactID = strings.Trim(strings.TrimSpace(artifactID), "/")
	if artifactID == "" {
		return nil, fmt.Errorf("artifact id is required")
	}
	note, err := s.fetchNote(ctx, ArtifactKey(artifactID))
	if err != nil {
		return nil, err
	}
	var artifact Artifact
	if err := json.Unmarshal([]byte(note.Body), &artifact); err != nil {
		return nil, fmt.Errorf("decode artifact metadata failed: %w", err)
	}
	if strings.TrimSpace(artifact.SourceURL) == "" {
		return nil, fmt.Errorf("artifact %q is missing sourceURL", artifactID)
	}
	if strings.TrimSpace(artifact.ArtifactID) == "" {
		artifact.ArtifactID = artifactID
	}
	return &artifact, nil
}

func (s *Service) OpenArtifact(ctx context.Context, artifactID string) (*Artifact, io.ReadCloser, error) {
	artifact, err := s.FetchArtifact(ctx, artifactID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.validateArtifact(artifact); err != nil {
		return nil, nil, err
	}
	reader, err := s.fs.OpenURL(ctx, artifact.SourceURL)
	if err != nil {
		return nil, nil, fmt.Errorf("read artifact source failed: %w", err)
	}
	return artifact, reader, nil
}

func (s *Service) fetchNote(ctx context.Context, key string) (*Note, error) {
	root, userID, err := s.ResolveRootURIContext(ctx)
	if err != nil {
		return nil, err
	}
	target := NoteURL(root, key)
	exists, err := s.fs.Exists(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("read scratchpad note failed: %w", err)
	}
	if !exists {
		return nil, &NoteNotFoundError{Key: key, URI: target}
	}
	note, err := s.readNoteURL(ctx, key, target)
	if err != nil {
		return nil, err
	}
	if note.Key != key {
		return nil, fmt.Errorf("scratchpad note identity mismatch for key %q", key)
	}
	if err := validateNoteUser(note, userID); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *Service) validateArtifact(artifact *Artifact) error {
	if artifact == nil {
		return fmt.Errorf("artifact metadata is required")
	}
	if artifact.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, artifact.ExpiresAt)
		if err == nil && !expiresAt.After(s.now().UTC()) {
			return fmt.Errorf("artifact %q is expired", artifact.ArtifactID)
		}
	}
	if len(s.allowedTargetSchemes) == 0 {
		return nil
	}
	scheme := strings.ToLower(afsurl.Scheme(artifact.SourceURL, "file"))
	if !s.allowedTargetSchemes[scheme] {
		return fmt.Errorf("artifact source scheme %q is not allowed", scheme)
	}
	return nil
}

func (s *Service) writeNote(ctx context.Context, root, key string, note Note) error {
	data, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		return err
	}
	target := NoteURL(root, key)
	if err = s.fs.Upload(ctx, target, 0o644, strings.NewReader(string(data)+"\n")); err != nil {
		return fmt.Errorf("write scratchpad note %q failed: %w", key, err)
	}
	return nil
}

func AppendNoteBody(existing, addition string) string {
	if strings.TrimSpace(existing) == "" {
		return addition
	}
	return strings.TrimRight(existing, "\n") + appendSeparator + strings.TrimLeft(addition, "\n")
}

func (s *Service) readNoteURL(ctx context.Context, key, uri string) (*Note, error) {
	data, err := s.fs.DownloadWithURL(ctx, uri)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &NoteNotFoundError{Key: key, URI: uri, Cause: err}
		}
		return nil, fmt.Errorf("read scratchpad note failed: %w", err)
	}
	var note Note
	if err = json.Unmarshal(data, &note); err != nil {
		return nil, fmt.Errorf("decode scratchpad note failed: %w", err)
	}
	if strings.TrimSpace(note.Key) == "" {
		return nil, fmt.Errorf("scratchpad note is missing key")
	}
	return &note, nil
}

func validateNoteUser(note *Note, userID string) error {
	if note == nil {
		return fmt.Errorf("scratchpad note is required")
	}
	noteUserID := strings.TrimSpace(note.UserID)
	if noteUserID == "" {
		return fmt.Errorf("scratchpad note %q is missing userId", note.Key)
	}
	if noteUserID != strings.TrimSpace(userID) {
		return fmt.Errorf("scratchpad note %q belongs to a different user", note.Key)
	}
	return nil
}

func (s *Service) ensureRoot(ctx context.Context, root string) error {
	if ok, err := s.fs.Exists(ctx, root); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else if ok {
		return nil
	}
	return s.fs.Create(ctx, root, 0o755, true)
}

func (s *Service) ResolveRootURI() (string, string, error) {
	return ResolveRootURI(s.rootTemplate, s.userID, s.basePath, s.macros)
}

func (s *Service) ResolveRootURIContext(ctx context.Context) (string, string, error) {
	return ResolveRootURI(s.rootTemplate, s.effectiveUserID(ctx), s.basePath, s.macros)
}

func (s *Service) effectiveUserID(ctx context.Context) string {
	if userID := UserIDFromContext(ctx); userID != "" {
		return userID
	}
	if s.userIDProvider != nil {
		if userID := strings.TrimSpace(s.userIDProvider(ctx)); userID != "" {
			return userID
		}
	}
	return strings.TrimSpace(s.userID)
}

func ResolveRootURI(template, userID, basePath string, macros map[string]string) (string, string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", "", fmt.Errorf("scratchpad requires an effective user id")
	}
	if strings.TrimSpace(template) == "" {
		template = DefaultRootURITemplate
	}
	if !strings.Contains(template, "${userID}") && !strings.Contains(template, "${user}") {
		return "", "", fmt.Errorf("scratchpad root template must include ${userID} or ${user}")
	}
	safeUserID := SanitizePathComponent(userID)
	if safeUserID == "" {
		return "", "", fmt.Errorf("effective user id cannot be converted to a scratchpad path segment")
	}
	out := strings.ReplaceAll(template, "${userID}", safeUserID)
	out = strings.ReplaceAll(out, "${user}", safeUserID)
	out = expandHostPathMacros(out, macros)
	out = strings.ReplaceAll(out, "${userID}", safeUserID)
	out = strings.ReplaceAll(out, "${user}", safeUserID)
	return normalizeRootURI(out, basePath), userID, nil
}

func expandHostPathMacros(value string, macros map[string]string) string {
	out := strings.TrimSpace(value)
	for key, value := range macros {
		out = strings.ReplaceAll(out, "${"+key+"}", value)
	}
	if strings.Contains(out, "${home}") {
		if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
			out = strings.ReplaceAll(out, "${home}", home)
		}
	}
	if strings.HasPrefix(out, "~/") || out == "~" {
		if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
			if out == "~" {
				out = home
			} else {
				out = filepath.Join(home, strings.TrimPrefix(out, "~/"))
			}
		}
	}
	return strings.TrimSpace(out)
}

func NoteURL(root, key string) string {
	sum := sha256.Sum256([]byte(key))
	return afsurl.Join(root, hex.EncodeToString(sum[:])+".json")
}

func ArtifactKey(artifactID string) string {
	return "artifact/" + strings.Trim(strings.TrimSpace(artifactID), "/")
}

var unsafePathComponent = regexp.MustCompile(`[^A-Za-z0-9_.@-]+`)

func SanitizePathComponent(value string) string {
	value = strings.TrimSpace(value)
	value = unsafePathComponent.ReplaceAllString(value, "_")
	value = strings.Trim(value, "._-")
	return value
}

func normalizeRootURI(root, basePath string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return root
	}
	if strings.Contains(root, "://") {
		return strings.TrimRight(root, "/")
	}
	if filepath.IsAbs(root) || isWindowsAbsPath(root) {
		return filepath.Clean(root)
	}
	if strings.TrimSpace(basePath) != "" {
		return filepath.Clean(filepath.Join(basePath, root))
	}
	return filepath.Clean(root)
}

func isWindowsAbsPath(v string) bool {
	if len(v) < 2 {
		return false
	}
	if v[1] != ':' {
		return false
	}
	return (v[0] >= 'a' && v[0] <= 'z') || (v[0] >= 'A' && v[0] <= 'Z')
}
