package gateway

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
)

type Gateway struct {
	server *echo.Echo
}

func NewGateway() *Gateway {
	gw := &Gateway{
		server: echo.New(),
	}
	gw.initRoutes()
	return gw
}

func (g *Gateway) initRoutes() {
	g.server.Any("/music/*", g.proxyToService("http://localhost:11000"))
	g.server.Any("/user/*", g.proxyToService("http://localhost:12000"))
	g.server.Any("/stats/*", g.proxyToService("http://localhost:13000"))
}

func (g *Gateway) proxyToService(target string) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Получаем только host без пути
		targetURL, err := url.Parse(target)
		if err != nil {
			return err
		}

		// Логируем путь, который перенаправляем
		fmt.Println("Proxying request to:", target+c.Request().URL.Path)

		// Создаем прокси-сервер
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Заменяем хост запроса на целевой
		c.Request().Host = targetURL.Host
		c.Request().URL.Scheme = targetURL.Scheme
		c.Request().URL.Host = targetURL.Host

		// Проксируем запрос
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func (g *Gateway) Start(port string) {
	g.server.Logger.Fatal(g.server.Start(port))
}
