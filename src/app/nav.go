package main

import "strings"

type navList []navItem

func (l *navList) Append(val navItem) {
	*l = append(*l, val)
}

func (l navList) Dup() (r navList) {
	r = make(navList, 0, len(l))
	for _, v := range l {
		r = append(r, v.Dup())
	}
	return
}

func (l navList) SetActive(path string) {
	for i, v := range l {
		if strings.HasPrefix(v.Href(), path) {
			l[i].AddClass("active")
		}
	}
}

type navItem interface {
	Title() string
	Href() string
	Class() string
	AddClass(string)
	IsDivider() bool
	IsHeader() bool
	Items() navList
	Dup() navItem
}

type navBase struct {
	title string
	href  string
	class []string
}

func (b *navBase) Title() string      { return b.title }
func (b *navBase) Href() string       { return b.href }
func (b *navBase) Class() string      { return strings.Join(b.class, " ") }
func (b *navBase) AddClass(cl string) { b.class = append(b.class, cl) }
func (b *navBase) IsDivider() bool    { return false }
func (b *navBase) IsHeader() bool     { return false }
func (b *navBase) Items() navList     { return nil }
func (b *navBase) Dup() navItem {
	var cl []string
	if b.class != nil {
		cl = make([]string, len(b.class))
		for i, v := range b.class {
			cl[i] = v
		}
	}
	return &navBase{
		title: b.title,
		href:  b.href,
		class: cl,
	}
}
