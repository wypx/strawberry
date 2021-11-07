package vm_admin

import (
	Global "vm_manager/vm_admin/global"
	Router "vm_manager/vm_admin/router"
	System "vm_manager/vm_admin/system"
)

/**
 * @Description:
 * @receiver:
 * @param:
 * @param:
 */
func Initialize() {
	env := Global.GetEnvironmentConfig()

	System.InitializeSystem(env)

	System.InitializeWebServer(env.WebServerAddr(), Router.InitializeRouter(env))
}
