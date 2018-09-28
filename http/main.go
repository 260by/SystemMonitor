package main

import (
	"flag"
	"time"
	"fmt"
	"os"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
	"github.com/gorilla/securecookie"
	"github.com/260by/SystemMonitor/data"
	"github.com/260by/tools/gconfig"
	"golang.org/x/crypto/bcrypt"
)

var (
	cookieName = "evenoteid"
	hashKey = []byte("the-big-and-secret-fash-key-here")
	blockKey = []byte("lot-secret-of-characters-big-too")
	secureCookie = securecookie.New(hashKey, blockKey)

	sess = sessions.New(sessions.Config{
		Cookie: cookieName,
		Encode: secureCookie.Encode,
		Decode: secureCookie.Decode,
		AllowReclaim: true,
		Expires: time.Minute * time.Duration(30),
	})
	maxNoteNum = 80

	user = data.User{}
	assets = data.Assets{}
	monitor = data.Monitor{}
)

// Config 配置
type Config struct {
	Database struct {
		Driver string
		Dsn string
		ShowSQL bool
		Migrate bool
	}
	HTTPServer struct {
		IP string
		Port int
	}
}

// 生成一个hashed密码
func generatePassword(userPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}

// 检查密码是否匹配。
func validatePassword(userPassword string, hashed []byte) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(hashed, []byte(userPassword)); err != nil {
		return false, err
	}
	return true, nil
}

func initAdminUser() (*data.User, error) {
	hashPassword, err := generatePassword("admin")
	if err != nil {
		return nil, err
	}
	user := new(data.User)
	user.UserName = "admin"
	user.Password = string(hashPassword)

	return user, nil
}

func authentication(ctx iris.Context) (auth bool, user string, err error) {
	s := sess.Start(ctx)
	auth, _ = s.GetBoolean("authenticated")
	user = s.GetString("userName")
	if err != nil {
		return
	}
	return
}

func main()  {
	var configFile = flag.String("config", "config.toml", "Configration file")
	var initUser = flag.Bool("initUser", false, "Init admin user")
	flag.Parse()

	app := iris.New()

	app.Logger().SetLevel("debug")
	app.Use(recover.New())
	app.Use(logger.New())

	var config = Config{}
	err := gconfig.Parse(*configFile, &config)
	if err != nil {
		app.Logger().Fatal(err)
	}

	orm, err := data.Connect(config.Database.Driver, config.Database.Dsn, config.Database.ShowSQL)
	// orm, err := xorm.NewEngine("mysql", "root:power123@tcp(192.168.1.251:3306)/evenote?charset=utf8")
	if err != nil {
		app.Logger().Fatalf("orm failed to connection: %v", err)
	}

	iris.RegisterOnInterrupt(func()  {
		orm.Close()
	})

	app.RegisterView(iris.HTML("./templates", ".html").Reload(true))	// 配置模板文件目录，并自动重新加载
	app.StaticWeb("/assets", "./assets")	// 配置静态文件目录，并映射为/assets

	app.Get("/", func(ctx iris.Context)  {
		ctx.Redirect("/login")
	})

	if *initUser {
		user, err := initAdminUser()
		if err != nil {
			panic(err)
		}
		_, err = orm.Insert(user)
		if err != nil {
			panic(err)
		}
		fmt.Println("Initialization admin user success.")
		os.Exit(0)
	}


	app.Get("/login", func(ctx iris.Context)  {
		s := sess.Start(ctx)
		auth, _ := s.GetBoolean("authenticated")
		if auth {
			ctx.Redirect("/evenote")
		}

		ctx.View("login.html")
	})

	app.Post("/login", func(ctx iris.Context)  {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		user := new(data.User)
		_, err := orm.Where("user_name = ?", username).Get(user)
		if err != nil {
			app.Logger().Fatalf("select user table failed: ", err)
		}

		hashPassword := user.Password
		result, err := validatePassword(password, []byte(hashPassword))

		if username == user.UserName && result {
			s := sess.Start(ctx)
			s.Set("userName", user.UserName)
			s.Set("authenticated", true)
			ctx.Redirect("/admin")
		} else {
			ctx.ViewData("valid", "用户名或密码错误")
			ctx.View("login.html")
		}
		
	})

	app.Get("/admin", func(ctx iris.Context)  {
		auth, user, err := authentication(ctx) // 认证函数，验证是否认证，认证成功获取用户ID
		if !auth {
			ctx.StatusCode(iris.StatusUnauthorized)
			return
		}
		if err != nil {
			app.Logger().Fatalf("Get sessions user id failed", err)
		}
		
		
		ctx.ViewData("user", user)
		ctx.View("admin.html")
	})

	app.Get("/logout", func(ctx iris.Context)  {
		sess.Start(ctx).Clear()
		ctx.Redirect("/login")
	})

	addr := fmt.Sprintf("%s:%v", config.HTTPServer.IP, config.HTTPServer.Port)
	app.Run(
		iris.Addr(addr),
		iris.WithoutServerError(iris.ErrServerClosed),
	)
}