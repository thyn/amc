package controllers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	as "github.com/aerospike/aerospike-client-go"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/citrusleaf/amc/common"
	"github.com/citrusleaf/amc/controllers/middleware/sessions"
	"github.com/citrusleaf/amc/models"
)

var (
	_observer            *models.ObserverT
	_defaultClientPolicy = as.NewClientPolicy()
)

func postSessionTerminate(c echo.Context) error {
	invalidateSession(c)
	return c.JSONBlob(http.StatusOK, []byte(`{"status": "success"}`))
}

func getDebug(c echo.Context) error {
	res := _observer.DebugStatus()
	if res.On {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":            "success",
			"debugging":         "ON",
			"initiator":         res.Initiator,
			"isOriginInitiator": res.Initiator == c.Request().RemoteAddr,
			"start_time":        res.StartTime.UnixNano() / 1e6, //1484923724160,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "success",
		"debugging": "OFF",
		"initiator": nil,
	})
}

func postDebug(c echo.Context) error {
	form := struct {
		Service      string `form:"service"`
		DurationMins int    `form:"duration"`
		Username     string `form:"username"`
	}{}

	c.Bind(&form)

	var res models.DebugStatus
	switch form.Service {
	case "start":
		res = _observer.StartDebug(c.Request().RemoteAddr, time.Duration(form.DurationMins)*time.Minute)
	case "restart":
		res = _observer.StartDebug(c.Request().RemoteAddr, time.Duration(form.DurationMins)*time.Minute)
	case "stop":
		res = _observer.StopDebug()
	default:
		return c.JSON(http.StatusOK, errorMap("Cluster not found"))
	}

	if res.On {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":     "success",
			"debugging":  "ON",
			"initiator":  res.Initiator,
			"start_time": res.StartTime.UnixNano() / 1e6, //1484923724160,
			"service":    form.Service,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "success",
		"debugging": "OFF",
		"initiator": nil,
		"service":   form.Service,
	})
}

func getAMCVersion(c echo.Context) error {
	return c.JSONBlob(http.StatusOK, []byte(fmt.Sprintf(`{"amc_version": "%s", "amc_type": "%s"}`, common.AMCVersion, common.AMCEdition)))
}

func Server(config *common.Config) {
	_observer = models.New(config)

	_defaultClientPolicy.Timeout = 10 * time.Second
	_defaultClientPolicy.LimitConnectionsToQueueSize = true
	_defaultClientPolicy.ConnectionQueueSize = 1

	e := echo.New()
	// Avoid stale connections
	e.ReadTimeout = 30 * time.Second
	e.WriteTimeout = 30 * time.Second

	store := sessions.NewCookieStore([]byte("amc-secret-key"))
	e.Use(sessions.Sessions("amc_session", store))

	if config.AMC.StaticPath == "" {
		log.Fatalln("No static dir has been set in the config file. Quiting...")
	}
	log.Infoln("Static files path is being set to:" + config.AMC.StaticPath)
	e.Static("/", config.AMC.StaticPath)
	e.Static("/static", config.AMC.StaticPath)

	// Middleware
	if !common.AMCIsProd() {
		// e.Logger.SetOutput(log.StandardLogger().Writer())
		// e.Use(middleware.Logger())
	} else {
		e.Use(middleware.Recover())
	}

	// Basic Authentication Middleware Setup
	basicAuthUser := os.Getenv("AMC_AUTH_USER")
	if basicAuthUser == "" {
		basicAuthUser = config.BasicAuth.User
	}

	basicAuthPassword := os.Getenv("AMC_AUTH_PASSWORD")
	if basicAuthPassword == "" {
		basicAuthPassword = config.BasicAuth.Password
	}

	if basicAuthUser != "" {
		e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) bool {
			if username == basicAuthUser && password == basicAuthPassword {
				return true
			}
			return false
		}))
	}

	// Routes
	e.POST("/session-terminate", postSessionTerminate)

	e.GET("/aerospike/service/debug", getDebug)
	e.POST("/aerospike/service/clusters/:clusterUuid/debug", postDebug) // cluster does not matter here

	e.GET("/alert-emails", getAlertEmails)
	e.POST("/alert-emails", postAlertEmails)
	e.POST("/delete-alert-emails", deleteAlertEmails)

	e.GET("/get_amc_version", getAMCVersion)
	e.GET("/get_current_monitoring_clusters", getCurrentMonitoringClusters)

	e.GET("/aerospike/get_multicluster_view/:port", getMultiClusterView)

	e.POST("/set-update-interval/:clusterUuid", setClusterUpdateInterval)
	e.GET("/aerospike/service/clusters/:clusterUuid", getCluster)
	e.POST("/aerospike/service/clusters/:clusterUuid/logout", postRemoveClusterFromSession)

	e.GET("/aerospike/service/clusters/:clusterUuid/udfs", getClusterUDFs)
	e.POST("/aerospike/service/clusters/:clusterUuid/drop_udf", postClusterDropUDF)
	e.POST("/aerospike/service/clusters/:clusterUuid/add_udf", postClusterAddUDF)

	e.POST("/aerospike/service/clusters/:clusterUuid/fire_cmd", postClusterFireCmd)
	e.GET("/aerospike/service/clusters/:clusterUuid/get_all_users", getClusterAllUsers)
	e.GET("/aerospike/service/clusters/:clusterUuid/get_all_roles", getClusterAllRoles)
	e.POST("/aerospike/service/clusters/:clusterUuid/add_user", postClusterAddUser)
	e.POST("/aerospike/service/clusters/:clusterUuid/user/:user/remove", postClusterDropUser)
	e.POST("/aerospike/service/clusters/:clusterUuid/user/:user/update", postClusterUpdateUser)
	e.POST("/aerospike/service/clusters/:clusterUuid/roles/:role/add_role", postClusterAddRole)
	e.POST("/aerospike/service/clusters/:clusterUuid/roles/:role/update", postClusterUpdateRole)
	e.POST("/aerospike/service/clusters/:clusterUuid/roles/:role/drop_role", postClusterDropRole)

	e.POST("/aerospike/service/clusters/:clusterUuid/initiate_backup", postInitiateBackup)
	e.GET("/aerospike/service/clusters/:clusterUuid/get_backup_progress", getBackupProgress)
	e.GET("/aerospike/service/clusters/:clusterUuid/get_successful_backups", getSuccessfulBackups)
	e.POST("/aerospike/service/clusters/:clusterUuid/get_available_backups", getAvailableBackups)

	e.POST("/aerospike/service/clusters/:clusterUuid/initiate_restore", postInitiateRestore)
	e.GET("/aerospike/service/clusters/:clusterUuid/get_restore_progress", getRestoreProgress)

	e.GET("/aerospike/service/clusters/:clusterUuid/throughput", getClusterThroughput)
	e.GET("/aerospike/service/clusters/:clusterUuid/throughput_history", getClusterThroughputHistory)
	e.GET("/aerospike/service/clusters/:clusterUuid/latency/:nodes", getNodeLatency)
	e.GET("/aerospike/service/clusters/:clusterUuid/latency_history/:nodes", getNodeLatencyHistory)
	e.GET("/aerospike/service/clusters/:clusterUuid/basic", getClusterBasic)
	e.POST("/aerospike/service/clusters/:clusterUuid/change_password", postClusterChangePassword)
	e.GET("/aerospike/service/clusters/:clusterUuid/alerts", getClusterAlerts)
	e.POST("/aerospike/service/clusters/:clusterUuid/add_node", postAddClusterNodes)
	e.GET("/aerospike/service/clusters/:clusterUuid/nodes/:nodes", getClusterNodes)
	e.GET("/aerospike/service/clusters/:clusterUuid/nodes/:node/allconfig", getClusterNodeAllConfig)
	e.POST("/aerospike/service/clusters/:clusterUuid/nodes/:nodes/setconfig", setClusterNodesConfig)
	e.POST("/aerospike/service/clusters/:clusterUuid/nodes/:node/switch_xdr_off", postSwitchXDROff)
	e.POST("/aerospike/service/clusters/:clusterUuid/nodes/:node/switch_xdr_on", postSwitchXDROn)
	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespaces", getClusterNamespaces)
	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/nodes/:nodes", getClusterNamespaceNodes)
	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/nodes/:node/allconfig", getClusterNamespaceAllConfig)
	e.POST("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/nodes/:node/setconfig", setClusterNamespaceConfig)
	e.GET("/aerospike/service/clusters/:clusterUuid/xdr/:xdrPort/nodes/:nodes", getClusterXdrNodes)
	e.GET("/aerospike/service/clusters/:clusterUuid/xdr/:xdrPort/nodes/:nodes/allconfig", getClusterXdrNodesAllConfig)
	e.POST("/aerospike/service/clusters/:clusterUuid/xdr/:xdrPort/nodes/:nodes/setconfig", setClusterXdrNodesConfig)
	e.GET("/aerospike/service/clusters/:clusterUuid/nodes/:node/allstats", getClusterNodeAllStats)
	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/nodes/:node/allstats", getClusterNamespaceNodeAllStats)
	e.GET("/aerospike/service/clusters/:clusterUuid/xdr/:port/nodes/:node/allstats", getClusterXdrNodeAllStats)

	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/sindexes/:sindex/nodes/:node/allstats", getClusterNamespaceSindexNodeAllStats)
	e.POST("/aerospike/service/clusters/:clusterUuid/namespace/:namespace/add_index", postClusterAddIndex)
	e.POST("/aerospike/service/clusters/:clusterUuid/namespace/:namespace/drop_index", postClusterDropIndex)

	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/sindexes", getClusterNamespaceSindexes)
	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/sets", getClusterNamespaceSets)
	e.GET("/aerospike/service/clusters/:clusterUuid/namespaces/:namespace/storage", getClusterNamespaceStorage)
	e.GET("/aerospike/service/clusters/:clusterUuid/jobs/nodes/:node", getClusterJobsNode)

	e.POST("/aerospike/service/clusters/get-cluster-id", postGetClusterId)
	e.GET("/aerospike/service/clusters/:clusterUuid/get-current-user", getClusterCurrentUser)
	e.GET("/aerospike/service/clusters/:clusterUuid/get_user_roles", getClusterUserRoles)

	log.Infof("Starting AMC server, version: %s %s", common.AMCVersion, common.AMCEdition)
	// Start server
	if config.AMC.CertFile != "" {
		log.Infof("In HTTPS (secure) Mode")
		e.StartTLS(config.AMC.Bind, config.AMC.CertFile, config.AMC.KeyFile)
	} else {
		log.Infof("In HTTP (insecure) Mode.")
		e.Start(config.AMC.Bind)
	}
}
