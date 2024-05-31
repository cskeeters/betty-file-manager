package main

import (
	"log"
	"io/fs"
)

// ByName implements sort.Interface for []DirEntry based on the Name() field.
type ByName []fs.DirEntry

// sort.Interface requires Len, Swap and Less
func (a ByName) Len() int {
	return len(a)
}

func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByName) Less(i, j int) bool {
	return a[i].Name() < a[j].Name()
}

// ByMod implements sort.Interface for []DirEntry based on the Name() field.
type ByMod []fs.DirEntry

// sort.Interface requires Len, Swap and Less
func (a ByMod) Len() int {
	return len(a)
}

func (a ByMod) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByMod) Less(i, j int) bool {
	iinfo, err := a[i].Info()
	if err != nil {
		log.Printf("Error getting ModTime of %s: %s", a[i].Name(), err)
		return true
	}
	jinfo, err := a[j].Info()
	if err != nil {
		log.Printf("Error getting ModTime of %s: %s", a[j].Name(), err)
		return false
	}
	// Last Modified First
	return iinfo.ModTime().After(jinfo.ModTime())
}

// BySize implements sort.Interface for []DirEntry based on the Name() field.
type BySize []fs.DirEntry

// sort.Interface requires Len, Swap and Less
func (a BySize) Len() int {
	return len(a)
}

func (a BySize) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a BySize) Less(i, j int) bool {
	iinfo, err := a[i].Info()
	if err != nil {
		log.Fatalf("Error getting ModTime of %s: %s", a[i].Name(), err)
	}
	jinfo, err := a[j].Info()
	if err != nil {
		log.Fatalf("Error getting ModTime of %s: %s", a[j].Name(), err)
	}
	return iinfo.Size() < jinfo.Size()
}


