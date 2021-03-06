package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7802150"
	"github.com/shurcooL/go/vcs"
)

// Godeps describes what a package needs to be rebuilt reproducibly.
// It's the same information stored in file Godeps.
type Godeps struct {
	ImportPath string
	GoVersion  string
	Packages   []string `json:",omitempty"` // Arguments to save, if any.
	Deps       []Dependency
}

// A Dependency is a specific revision of a package.
type Dependency struct {
	ImportPath string
	Comment    string `json:",omitempty"` // Description of commit, if present.
	Rev        string // VCS-specific commit ID.
}

func ReadGodeps(path string, g *Godeps) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(g)
}

// ---

type GoPackagesFromGodeps struct {
	path string

	Entries []*gist7480523.GoPackage

	gist7802150.DepNode2
}

func NewGoPackagesFromGodeps(path string) *GoPackagesFromGodeps {
	return &GoPackagesFromGodeps{path: path}
}

func (this *GoPackagesFromGodeps) Update() {
	// TODO: Have a source?

	g := Godeps{}
	err := ReadGodeps(this.path, &g)
	if err != nil {
		log.Fatalln("ReadGodeps:", err)
	}

	this.Entries = nil
	for _, dependency := range g.Deps {
		if goPackage := gist7480523.GoPackageFromImportPath(dependency.ImportPath); goPackage != nil {
			// TODO: Can try to optimize by not blocking on UpdateVcs() here...
			gist7802150.MakeUpdatedLock.Unlock() // HACK: Needed because UpdateVcs() calls MakeUpdated().
			goPackage.UpdateVcs()
			gist7802150.MakeUpdatedLock.Lock() // HACK
			if goPackage.Dir.Repo == nil {
				continue
			}
			goPackage.Dir.Repo.Vcs = &FixedLocalRevVcs{LocalRev: dependency.Rev, Vcs: goPackage.Dir.Repo.Vcs}

			this.Entries = append(this.Entries, goPackage)
		}
	}
}

func (this *GoPackagesFromGodeps) List() []*gist7480523.GoPackage {
	return this.Entries
}

// ---

type FixedLocalRevVcs struct {
	vcs.Vcs

	LocalRev string
}

func (this *FixedLocalRevVcs) GetLocalRev() string {
	return this.LocalRev
}

func (this *FixedLocalRevVcs) IsContained(rev string) bool {
	// HACK? This is needed so that it consideres all different remote versions as updates (instead of needing to push).
	return false
}
