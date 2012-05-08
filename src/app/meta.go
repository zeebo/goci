package main

type Meta struct {
	CSS       list
	JS        list
	BaseTitle string
	Title     string
}

func (m *Meta) Dup() *Meta {
	return &Meta{
		CSS:       m.CSS.Dup(),
		JS:        m.JS.Dup(),
		Title:     m.Title,
		BaseTitle: m.BaseTitle,
	}
}
