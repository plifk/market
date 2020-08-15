package frontend

import (
	"net/http"

	"github.com/plifk/market/internal/services"
)

// AdminHandler for the application.
type AdminHandler struct {
	Frontend *Frontend

	dashboardHandler *AdminDashboardHandler
	usersHandler     *AdminUsersHandler
	securityHandler  *AdminSecurityHandler
	reportsHandler   *AdminReportsHandler
	rolesHandler     *AdminRolesHandler
}

// Load /admin routes.
func (h *AdminHandler) Load() {
	h.dashboardHandler = &AdminDashboardHandler{Frontend: h.Frontend}
	h.usersHandler = &AdminUsersHandler{Frontend: h.Frontend}
	h.securityHandler = &AdminSecurityHandler{Frontend: h.Frontend}
	h.reportsHandler = &AdminReportsHandler{Frontend: h.Frontend}
	h.rolesHandler = &AdminRolesHandler{Frontend: h.Frontend}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if user := services.UserFromRequest(r); user.Access != services.AdminAuthorization {
		h.Frontend.HTTPError(w, r, http.StatusNotFound)
		return
	}

	var handler http.Handler
	switch route := dirRouter(r.URL.Path); {
	case route.is("/admin"):
		handler = h.dashboardHandler
	case route.is("/admin/users"):
		handler = h.usersHandler
	case route.is("/admin/security"):
		handler = h.securityHandler
	case route.is("/admin/reports"):
		handler = h.reportsHandler
	case route.is("/admin/roles"):
		handler = h.rolesHandler
	}
	if handler == nil {
		handler = h.Frontend.staticHandler
	}
	handler.ServeHTTP(w, r)
}

// AdminDashboardHandler for the application.
type AdminDashboardHandler struct {
	Frontend *Frontend
}

// DashboardHandler for /admin.
func (h *AdminDashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// AdminUsersHandler for the application.
type AdminUsersHandler struct {
	Frontend *Frontend
}

// UsersHandler for /admin/users.
func (h *AdminUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// AdminSecurityHandler for the application.
type AdminSecurityHandler struct {
	Frontend *Frontend
}

// UsersHandler for /admin/security.
func (h *AdminSecurityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// AdminReportsHandler for the application.
type AdminReportsHandler struct {
	Frontend *Frontend
}

// UsersHandler for /admin/reports.
func (h *AdminReportsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// AdminRolesHandler for the application.
type AdminRolesHandler struct {
	Frontend *Frontend
}

// UsersHandler for /admin/roles.
func (h *AdminRolesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
