package vo

import (
	"github.com/ARUMANDESU/goread/backend/pkg/epubx"
)

type Metadata struct {
	Title       string
	Authors     []string
	Publishers  []string
	Date        string
	Languages   []string
	Subjects    []string
	Description string
	ISBN        string
}

func MetadataFromEPUB(em epubx.Metadata) Metadata {
	var m Metadata
	if len(em.Titles) > 0 && em.Titles[0] != "" {
		m.Title = em.Titles[0]
	}
	if len(em.Dates) > 0 && em.Dates[0] != "" {
		m.Date = em.Dates[0]
	}

	for _, v := range em.Creators {
		m.Authors = append(m.Authors, v.Name)
	}
	for _, v := range em.Identifiers {
		if v.Scheme == "ISBN" {
			m.ISBN = v.ID
		}
	}
	m.Publishers = em.Publishers
	m.Languages = em.Languages
	m.Subjects = em.Subjects
	m.Description = em.Description

	return m
}
