# Caller-scoped resource roots

`resource.NewProvider` supplies ordinary AFS files under a caller-defined scheme.
Each authority names one explicitly configured, exact root. It never interprets
an authority as a hostname to discover or fetch. Application metadata, skill
frontmatter, MCP schemas and list/get protocol methods belong to their callers.

```go
//go:embed skills
var skills embed.FS

backend := afs.New() // caller owns credentials, native providers and CloseAll
provider, err := resource.NewProvider(resource.Config{
    Scheme: "skill",
    Service: backend,
    Roots: map[string]resource.Root{
        "datly-reader": {FS: skills, Path: "skills/datly-reader"},
        "local":        {URL: "file://localhost/absolute/existing/directory"},
        "draft":        {URL: "mem://localhost/my-application/skills"},
    },
})
if err != nil { return err }
files := afs.NewWithProviders(map[string]afs.Provider{"skill": provider})
body, err := files.DownloadWithURL(ctx, "skill://datly-reader/SKILL.md")
// skill://datly-reader/references/guide.md uses the same exact root.
```

Create another provider for another scheme and put both in the service's map.
Create another service for another caller's mappings, even when scheme and
authority names are identical. Both maps are copied at construction; request
options cannot replace resource roots. There is no global `Register` call.
Unspecified service schemes still use AFS's existing global bootstrap registry;
this is a resource mapping API, not a service-wide allowlist. The caller remains
responsible for concurrency safety of borrowed `fs.FS` and backend services.
Closing the mapping does not close those borrowed resources.

## Operations

| Operation | fs.FS / embed.FS | URL root |
| --- | --- | --- |
| Object / metadata | fs.Stat | native AFS Object |
| List | fs.ReadDir, self + immediate children, native match/page options | native AFS List, native ordering/filtering |
| Open / Download | fs.FS.Open | native AFS OpenURL |
| Exists | fs.ErrNotExist becomes false | native AFS Exists |
| Upload / Create / Delete | ErrReadOnly | native AFS operation |
| NewWriter | ErrReadOnly | native AFS NewWriter, including its fallback |
| Walk / batch upload / Copy / Move | AFS service composition | AFS service composition |

File listings contain the file itself. Object and list URLs remain virtual;
metadata, including native size, mode, time and Unwrap, is preserved. For example,
the current mem provider reports size zero; mapping does not invent a new size.
Native URL-root errors are propagated, including provider-specific missing-path
errors. `errors.Is(err, resource.ErrUnknownNamespace)` distinguishes an unmapped
authority, without trying another root. `ErrBoundary` covers invalid paths and
rejected links. `ErrReadOnly` distinguishes filesystem-backed write attempts.

Copy/Move retain AFS streaming/tree semantics, including copying from an embedded
root into a writable root. No mapped Copier/Mover shortcut is advertised: native
server-side copy/rename optimization is not forwarded. AFS copy/move are not
atomic or transactional; moving from a read-only root can copy the destination
before failing to delete the source. No ACL, symlink-creation, or mapped ErrorCoder
capability is added. Native credentials and ordinary operation options are passed
to the backing service; a read Modifier runs once at the outer service boundary.
Backend-specific options carrying additional paths require caller review; options
are trusted configuration, not client input. Known native object.Link options are
rejected because they can address a second pathname.

## Boundaries and cancellation

Virtual paths are decoded once. Dot segments, empty interior segments, encoded
separators, backslashes, NUL, remaining percent escapes, and colon-containing path
segments are rejected. Query, fragment, userinfo and opaque URIs are rejected.
Spaces and Unicode names are supported; listed child names are URL-escaped.
Authorities match exactly, with no default namespace or implicit root search.
Root configuration also rejects traversal. The mapping delegates only to the
configured backend root; returned list URLs outside it are rejected.

For a local URL root, construction resolves the existing root's symlinks, and each
operation checks every existing target/ancestor component with Lstat. Symlinks
below that root, including links to another in-root file, are rejected. Directory
listings containing symlinks fail explicitly. Missing tails are allowed for writes.
These checks protect against stable symlink escapes, but native AFS file operations
remain pathname based: an external process can replace a component between the
check and the operation. This is **not race-resistant filesystem confinement**.
Use roots whose filesystem tree is controlled by the application. Supporting
untrusted concurrent filesystem writers requires a separate native file-provider
change using descriptor-relative operations; it is not claimed by this patch.

An arbitrary fs.FS must supply its own confinement; embed.FS is suitable. The
fs.FS interface cannot inspect the implementation's symlink resolution, so
os.DirFS/fs.Sub(os.DirFS(...)) alone is not a confinement mechanism. Local files
should use a file URL to receive the explicit checks above. FS configuration is
lexically validated; its root is opened lazily and missing roots fail normally.

Cancellation is checked before resolution, after metadata/list/open calls and
between reader/writer calls. Context is forwarded to native URL operations.
An already blocked fs.FS read cannot be interrupted through the fs.FS interface;
blocking I/O cancellation remains the backend's responsibility. Close still
releases returned streams after cancellation.

Cloud URLs can be configured through an existing caller-owned AFS service and its
native providers. This candidate has been exercised with embed.FS, fstest.MapFS,
local file URLs and native mem URLs only. No cloud connectivity, permission,
write, or server-side capability claim is made.

## Scratchpad

Scratchpad continues to own contextual user roots, note envelopes, artifact
records and allowed target schemes. Those policies are not plain-file mechanics
and have not been extracted or rewritten. The reusable native mechanism is the
service-scoped provider factory seam; a scratchpad regression test exercises it
while retaining contextual note isolation and artifact policy.
