package main

import "github.com/wise2c-dev/pagoda/database"
import "github.com/wise2c-dev/pagoda/playbook"

type Component struct {
	database.MetaComponent
	Hosts map[string][]*database.Host `json:"hosts"`
}

func NewComponent(clusterID string, cp *database.Component) *Component {
	c := &Component{
		MetaComponent: cp.MetaComponent,
		Hosts:         playbook.ConvertHosts(clusterID, cp.Hosts),
	}

	return c
}
