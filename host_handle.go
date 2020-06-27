package main

import (
	"net/http"

	"github.com/wise2c-dev/pagoda/database"

	"github.com/gin-gonic/gin"
)

func retrieveHosts(c *gin.Context) {
	clusterID := c.Param("cluster_id")

	hs, err := database.Instance().RetrieveHosts(clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, hs)
	}
}

func createHost(c *gin.Context) {
	clusterID := c.Param("cluster_id")

	h := &database.Host{}
	if err := c.BindJSON(h); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	err := database.Instance().CreateHost(clusterID, h)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, h)
	}
}

func deleteHost(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	hostID := c.Param("host_id")

	err := database.Instance().DeleteHost(clusterID, hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.Status(http.StatusOK)
	}
}

func updateHost(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	hostID := c.Param("host_id")

	h := &database.Host{}
	if err := c.BindJSON(h); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if h.ID != hostID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "two host id must be equal",
		})
		return
	}

	err := database.Instance().UpdateHost(clusterID, h)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, h)
	}
}

func retrieveHost(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	hostID := c.Param("host_id")

	h, err := database.Instance().RetrieveHost(clusterID, hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, h)
	}
}
