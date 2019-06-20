package database

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/go-xorm/xorm"
	"github.com/golang/glog"
	//just init
	_ "github.com/mattn/go-sqlite3"
)

type SQLEngine struct {
	xe *xorm.Engine
}

type EngineConfig struct {
	SQLType      string
	ShowSQL      bool
	ShowExecTime bool
	DBPath       string
	InitSQL      string
}

const (
	sqlType = "sqlite3"
)

var (
	i    *SQLEngine
	once = &sync.Once{}

	//DBPath - sqlite db file path
	DBPath *string
	//InitSQL - sql file to init db
	InitSQL *string
)

func Instance() *SQLEngine {
	once.Do(instance)

	return i
}

func instance() {
	sqlConfig := &EngineConfig{
		SQLType:      sqlType,
		ShowSQL:      false,
		ShowExecTime: true,
		DBPath:       *DBPath,
		InitSQL:      *InitSQL,
	}

	var err error
	i, err = newEngine(sqlConfig)
	if err != nil {
		panic(fmt.Sprintf("get sql engine instance error %v", err))
	}

	glog.V(1).Infof(
		"Create SQL Engine Instance with [sqlType: %s, dbPath: %s, initSQL: %s]",
		sqlConfig.SQLType,
		sqlConfig.DBPath,
		sqlConfig.InitSQL,
	)
}

func newEngine(config *EngineConfig) (*SQLEngine, error) {
	e := &SQLEngine{}
	engine, err := xorm.NewEngine(config.SQLType, config.DBPath)
	if err != nil {
		return nil, err
	}
	e.xe = engine

	e.xe.ShowSQL(config.ShowSQL)
	e.xe.ShowExecTime(config.ShowExecTime)
	_, err = e.xe.ImportFile(config.InitSQL)
	return e, err
}

func (e *SQLEngine) RetrieveClusters() ([]*Cluster, error) {
	clusters := make([]*Cluster, 0)
	err := e.xe.Find(&clusters)
	return clusters, err
}

func (e *SQLEngine) CreateCluster(c *Cluster) error {
	c.ID = newUUID()
	c.State = Initial

	if _, err := e.xe.InsertOne(c); err != nil {
		return err
	}

	return nil
}

func (e *SQLEngine) DeleteCluster(id string) error {
	var err error
	write := func(statement string, id string) sql.Result {
		if err != nil {
			return nil
		}
		var rs sql.Result
		rs, err = e.xe.Exec(statement, id)
		return rs
	}

	write("delete from cluster_component where cluster_id = ?", id)
	write("delete from cluster_host where cluster_id = ?", id)
	rs := write("delete from cluster where id = ?", id)

	if err != nil {
		return err
	}

	if i, err := rs.RowsAffected(); i == 0 || err != nil {
		return fmt.Errorf("can't find cluster %s", id)
	}

	return nil
}

func (e *SQLEngine) UpdateCluster(c *Cluster) error {
	i, err := e.xe.Id(c.ID).Update(c)
	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("can't find cluster %s", c.ID)
	}

	return nil
}

func (e *SQLEngine) RetrieveCluster(id string) (c *Cluster, err error) {
	c = &Cluster{}
	has, err := e.xe.ID(id).Get(c)

	if err != nil {
		glog.V(3).Info(err)
		return
	}

	if !has {
		err = fmt.Errorf("can't find cluster %s", id)
		return
	}

	cs, err := e.RetrieveComponents(id)
	if err != nil {
		glog.V(3).Info(err)
		return
	}
	c.Components = cs

	hs, err := e.RetrieveHosts(id)
	if err != nil {
		glog.V(3).Info(err)
		return
	}
	c.Hosts = hs

	return
}

func (e *SQLEngine) RetrieveComponents(
	clusterID string,
) (cs []*Component, err error) {
	ccs := make([]*ClusterComponent, 0)
	if err = e.xe.Where("cluster_id = ?", clusterID).Find(&ccs); err != nil {
		return
	}

	cs = make([]*Component, 0, len(ccs))
	for _, cc := range ccs {
		cs = append(cs, cc.Component)
	}
	return
}

func (e *SQLEngine) CreateComponent(clusterID string, cp *Component) error {
	cp.ID = newUUID()
	cc := &ClusterComponent{
		ClusterID:     clusterID,
		ComponentID:   cp.ID,
		ComponentName: cp.Name,
		Component:     cp,
	}

	_, err := e.xe.InsertOne(cc)
	return err
}

func (e *SQLEngine) DeleteComponent(clusterID string, id string) error {
	rs, err := e.xe.Exec(
		"delete from cluster_component where cluster_id = ? and component_id = ?",
		clusterID,
		id,
	)

	if err != nil {
		return err
	}

	if i, err := rs.RowsAffected(); i == 0 || err != nil {
		return fmt.Errorf("can't find component %s", id)
	}

	return nil
}

func (e *SQLEngine) UpdateComponent(clusterID string, cp *Component) error {
	ccp := &ClusterComponent{
		ClusterID:     clusterID,
		ComponentID:   cp.ID,
		ComponentName: cp.Name,
		Component:     cp,
	}
	i, err := e.xe.Where(
		"cluster_id = ? and component_id = ?",
		clusterID,
		cp.ID,
	).Update(ccp)

	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("can't find cluster %s component %s", clusterID, cp.Name)
	}

	return nil
}

func (e *SQLEngine) RetrieveComponent(
	clusterID string,
	id string,
) (*Component, error) {
	cc := &ClusterComponent{
		ClusterID:   clusterID,
		ComponentID: id,
		Component:   &Component{},
	}

	has, err := e.xe.Get(cc)
	if err != nil {
		return cc.Component, err
	} else if !has {
		return cc.Component, fmt.Errorf("can't find component %s", id)
	} else {
		return cc.Component, nil
	}
}

func (e *SQLEngine) RetrieveHosts(
	clusterID string,
) (hs []*Host, err error) {
	chs := make([]*ClusterHost, 0)
	if err = e.xe.Where("cluster_Id = ?", clusterID).Find(&chs); err != nil {
		return
	}

	hs = make([]*Host, 0, len(hs))
	for _, ch := range chs {
		hs = append(hs, ch.Host)
	}
	return
}

func (e *SQLEngine) CreateHost(clusterID string, h *Host) error {
	h.ID = newUUID()
	ch := &ClusterHost{
		ClusterID: clusterID,
		HostID:    h.ID,
		IP:        h.IP,
		Hostname:  h.HostName,
		Host:      h,
	}

	_, err := e.xe.InsertOne(ch)
	return err
}

func (e *SQLEngine) DeleteHost(clusterID string, hostID string) error {
	//check if there has components are running on host, if is,never delete host
	var components []*ClusterComponent
	var err error
	err = e.xe.Where("cluster_id = ?",clusterID).Find(&components)
	if err !=nil {
		return fmt.Errorf("get cluster components error -> %s", err.Error())
	}

	var hostHasComponent bool
	if len(components) > 0 {
		for _,v:=range components {
			for _,vv:=range v.Component.Hosts {
				for _,vvv := range vv {
					if vvv == hostID {
						hostHasComponent = true
						goto checkHostHasComponent
					}
				}
			}
		}
	}
checkHostHasComponent:
	if hostHasComponent {
		return fmt.Errorf("host=%s in cluster=%s has running components,so host could not be deleted", hostID,clusterID)
	}

	_, err = e.xe.Exec(
		"delete from cluster_host where cluster_id = ? and host_id = ?",
		clusterID,
		hostID,
	)
	return err
}

func (e *SQLEngine) UpdateHost(clusterID string, h *Host) error {
	ch := &ClusterHost{
		ClusterID: clusterID,
		HostID:    h.ID,
		IP:        h.IP,
		Hostname:  h.HostName,
		Host:      h,
	}
	i, err := e.xe.Where(
		"cluster_id = ? and host_id = ?",
		clusterID,
		ch.HostID,
	).Update(ch)

	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("can't find cluster %s host %s", clusterID, h.ID)
	}

	return nil
}

func (e *SQLEngine) RetrieveHost(
	clusterID string,
	hostID string,
) (*Host, error) {
	ch := &ClusterHost{
		ClusterID: clusterID,
		HostID:    hostID,
	}
	has, err := e.xe.Get(ch)

	if !has {
		return nil, fmt.Errorf("can't find cluster %s host %s", clusterID, hostID)
	}

	return ch.Host, err
}

func (e *SQLEngine) RetrieveLogs(clusterID string) ([]*Notification, error) {
	cls := make([]*ClusterLog, 0)

	if err := e.xe.
		Where("cluster_id = ?", clusterID).
		OrderBy("created").
		Find(&cls); err != nil {
		return nil, err
	}

	logs := make([]*Notification, 0, len(cls))
	for _, cl := range cls {
		logs = append(logs, cl.Log)
	}

	return logs, nil
}

func (e *SQLEngine) CreateLog(clusterID string, n *Notification) error {
	cn := &ClusterLog{
		ClusterID: clusterID,
		Log:       n,
	}

	_, err := e.xe.InsertOne(cn)
	return err
}

func (e *SQLEngine) DeleteLogs(clusterID string) error {
	_, err := e.xe.Exec(
		"delete from cluster_log where cluster_id = ?",
		clusterID,
	)
	return err
}
