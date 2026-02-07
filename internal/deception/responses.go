package deception

import "fmt"

// NextJSPage returns a minimal HTML page that looks like a
// Next.js application. The title and nonce vary to appear dynamic.
func NextJSPage(title, nonce string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charSet="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>%s</title>
<meta name="next-head-count" content="3"/>
<link rel="preload" href="/_next/static/css/app.css" as="style" nonce="%s"/>
<link rel="preload" href="/_next/static/chunks/main-app.js" as="script" nonce="%s"/>
</head>
<body>
<div id="__next"><div class="container"><main></main></div></div>
<script src="/_next/static/chunks/webpack.js" nonce="%s" async=""></script>
<script src="/_next/static/chunks/main-app.js" nonce="%s" async=""></script>
<script src="/_next/static/buildManifest.js" nonce="%s" async=""></script>
</body>
</html>`, title, nonce, nonce, nonce, nonce, nonce)
}

// NextJSErrorPage returns a Next.js-style error page.
func NextJSErrorPage(status int) string {
	msg := "Internal Server Error"
	if status == 404 {
		msg = "This page could not be found"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charSet="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>%d: %s</title>
<meta name="next-head-count" content="3"/>
<style>body{color:#000;background:#fff;margin:0}.next-error-h1{border-right:1px solid rgba(0,0,0,.3)}@media (prefers-color-scheme:dark){body{color:#fff;background:#000}.next-error-h1{border-right:1px solid rgba(255,255,255,.3)}}</style>
</head>
<body>
<div id="__next">
<div style="font-family:system-ui;height:100vh;text-align:center;display:flex;flex-direction:column;align-items:center;justify-content:center">
<div><style>body{color:#000;background:#fff;margin:0}</style>
<h1 class="next-error-h1" style="display:inline-block;margin:0 20px 0 0;padding:0 23px 0 0;font-size:24px;font-weight:500;vertical-align:top;line-height:49px">%d</h1>
<div style="display:inline-block"><h2 style="font-size:14px;font-weight:400;line-height:49px;margin:0">%s.</h2></div>
</div></div></div>
</body>
</html>`, status, msg, status, msg)
}

// NextJSRSCPayload returns a fake React Server Component flight
// response. This mimics the RSC wire format used by Next.js.
func NextJSRSCPayload() string {
	return `1:I["(app-pages-browser)/./src/app/layout.tsx",["app/layout","static/chunks/app/layout.js"],"default"]
2:I["(app-pages-browser)/./src/app/page.tsx",["app/page","static/chunks/app/page.js"],"default"]
0:["$","$L1",null,{"children":["$","$L2",null,{}]}]`
}

// NextJSServerActionResponse returns a fake server action response
// as would be returned by a Next.js app processing a form submission.
func NextJSServerActionResponse() string {
	return `{"actionResult":"$undefined","redirectURL":null}`
}

// NextJSBuildManifest returns a fake _buildManifest.js that
// references plausible chunk files.
func NextJSBuildManifest() string {
	return `self.__BUILD_MANIFEST={
  "polyfillFiles":["static/chunks/polyfills.js"],
  "devFiles":[],
  "ampDevFiles":[],
  "lowPriorityFiles":["static/css/app.css"],
  "rootMainFiles":["static/chunks/webpack.js","static/chunks/main-app.js"],
  "pages":{"/":["static/chunks/pages/index.js"]},
  "ampFirstPages":[]
};self.__BUILD_MANIFEST_CB&&self.__BUILD_MANIFEST_CB()`
}

// WordPressLoginPage returns a fake wp-login.php page.
func WordPressLoginPage(version string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en-US">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
<title>Log In &lsaquo; WordPress</title>
<meta name="robots" content="max-image-preview:large, noindex, noarchive"/>
<link rel="stylesheet" href="/wp-admin/css/login.min.css?ver=%s" type="text/css"/>
</head>
<body class="login login-action-login wp-core-ui">
<div id="login">
<h1><a href="https://wordpress.org/">Powered by WordPress</a></h1>
<form name="loginform" id="loginform" action="/wp-login.php" method="post">
<p><label for="user_login">Username or Email Address</label>
<input type="text" name="log" id="user_login" class="input" size="20" autocapitalize="off"/></p>
<p><label for="user_pass">Password</label>
<input type="password" name="pwd" id="user_pass" class="input" size="20"/></p>
<p class="submit">
<input type="submit" name="wp-submit" id="wp-submit" class="button button-primary button-large" value="Log In"/>
</p>
</form>
</div>
<div class="language-switcher"></div>
</body>
</html>`, version)
}

// PhpMyAdminLoginPage returns a fake phpMyAdmin login page.
func PhpMyAdminLoginPage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>phpMyAdmin</title>
<link rel="icon" href="/favicon.ico" type="image/x-icon"/>
<link rel="stylesheet" type="text/css" href="/themes/pmahomme/css/theme.css"/>
<style>body{font-family:sans-serif;background:#e7e7e7}#page_content{width:60%;margin:auto;background:#fff;padding:2em;margin-top:5em;border:1px solid #ddd;border-radius:5px}h1{color:#333}input[type=text],input[type=password]{width:100%;padding:8px;margin:4px 0 12px;border:1px solid #ccc;border-radius:3px;box-sizing:border-box}input[type=submit]{background:#f60;color:#fff;padding:10px 20px;border:none;border-radius:3px;cursor:pointer}.logo{text-align:center;margin-bottom:1em}</style>
</head>
<body>
<div id="page_content">
<div class="logo"><h1>phpMyAdmin</h1><p>5.2.1</p></div>
<form method="post" action="/phpmyadmin/" class="login">
<fieldset>
<legend>Log in</legend>
<label for="input_username">Username:</label>
<input type="text" name="pma_username" id="input_username" value="" autocomplete="username"/>
<label for="input_password">Password:</label>
<input type="password" name="pma_password" id="input_password" value="" autocomplete="current-password"/>
<label for="select_server">Server Choice:</label>
<select name="server" id="select_server"><option value="1">127.0.0.1</option></select>
</fieldset>
<fieldset class="tblFooters">
<input type="submit" id="input_go" value="Go"/>
</fieldset>
</form>
</div>
</body>
</html>`
}

// AdminerLoginPage returns a fake Adminer login page.
func AdminerLoginPage() string {
	return `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="robots" content="noindex"/>
<title>Login - Adminer</title>
<style>body{font-family:"Helvetica Neue",Helvetica,Arial,sans-serif;background:#e8ecef;color:#333}#content{max-width:30em;margin:5em auto;background:#fff;padding:2em;border:1px solid #ccc;border-radius:4px}h1{font-size:1.5em}table{width:100%}th{text-align:right;padding:.5em;width:30%}td{padding:.5em}input[type=text],input[type=password],select{width:100%;padding:6px;border:1px solid #bbb;border-radius:3px;box-sizing:border-box}input[type=submit]{background:#4479BA;color:#fff;padding:8px 20px;border:none;border-radius:3px;cursor:pointer}.version{text-align:center;color:#999;font-size:.85em}</style>
</head>
<body class="ltr nojs">
<div id="content">
<h1>Adminer</h1>
<form action="/adminer.php" method="post">
<table>
<tr><th>System:<td><select name="auth[driver]"><option value="server">MySQL</option><option value="pgsql">PostgreSQL</option><option value="sqlite">SQLite 3</option></select>
<tr><th>Server:<td><input type="text" name="auth[server]" value="localhost" autocapitalize="off"/>
<tr><th>Username:<td><input type="text" name="auth[username]" id="username" autocomplete="username" autocapitalize="off"/>
<tr><th>Password:<td><input type="password" name="auth[password]" autocomplete="current-password"/>
<tr><th>Database:<td><input type="text" name="auth[db]" autocapitalize="off"/>
</table>
<p><input type="submit" value="Login"/>
</form>
<p class="version">Adminer 4.8.1</p>
</div>
</body>
</html>`
}

// CPanelLoginPage returns a fake cPanel login page.
func CPanelLoginPage() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8"/>
<title>cPanel Login</title>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<style>*{margin:0;padding:0;box-sizing:border-box}body{font-family:Roboto,Arial,sans-serif;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);min-height:100vh;display:flex;align-items:center;justify-content:center}.login-container{background:#fff;border-radius:8px;box-shadow:0 10px 40px rgba(0,0,0,.2);padding:2em;width:350px}.logo{text-align:center;margin-bottom:1.5em}.logo h1{font-size:1.5em;color:#333}.logo h1 span{color:#ff6c2c}label{display:block;margin-bottom:.3em;color:#555;font-size:.9em}input[type=text],input[type=password]{width:100%;padding:10px;margin-bottom:1em;border:1px solid #ddd;border-radius:4px;font-size:1em}input[type=submit]{width:100%;padding:10px;background:#ff6c2c;color:#fff;border:none;border-radius:4px;font-size:1em;cursor:pointer}input[type=submit]:hover{background:#e55b1e}.footer{text-align:center;margin-top:1em;color:#999;font-size:.8em}</style>
</head>
<body>
<div class="login-container">
<div class="logo"><h1><span>c</span>Panel</h1></div>
<form action="/cpanel" method="post">
<label for="user">Username</label>
<input type="text" name="user" id="user" autocomplete="username"/>
<label for="pass">Password</label>
<input type="password" name="pass" id="pass" autocomplete="current-password"/>
<input type="submit" value="Log in"/>
</form>
<div class="footer">cPanel, Inc. Copyright 2024</div>
</div>
</body>
</html>`
}

// GenericAdminLoginPage returns a generic admin panel login page.
func GenericAdminLoginPage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>Admin Login</title>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<style>*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f5f5f5;min-height:100vh;display:flex;align-items:center;justify-content:center}.login-box{background:#fff;border-radius:6px;box-shadow:0 2px 10px rgba(0,0,0,.1);padding:2em;width:320px}h2{text-align:center;margin-bottom:1.5em;color:#333}label{display:block;margin-bottom:.3em;color:#555;font-size:.9em}input[type=text],input[type=password]{width:100%;padding:10px;margin-bottom:1em;border:1px solid #ddd;border-radius:4px}input[type=submit]{width:100%;padding:10px;background:#007bff;color:#fff;border:none;border-radius:4px;cursor:pointer}input[type=submit]:hover{background:#0056b3}</style>
</head>
<body>
<div class="login-box">
<h2>Admin Panel</h2>
<form action="/admin/login" method="post">
<label for="username">Username</label>
<input type="text" name="username" id="username" autocomplete="username"/>
<label for="password">Password</label>
<input type="password" name="password" id="password" autocomplete="current-password"/>
<input type="submit" value="Sign In"/>
</form>
</div>
</body>
</html>`
}

// SolrAdminPage returns a fake Apache Solr admin cores JSON response.
func SolrAdminPage() string {
	return `{
  "responseHeader":{
    "status":0,
    "QTime":2
  },
  "initFailures":{},
  "status":{
    "core0":{
      "name":"core0",
      "instanceDir":"/var/solr/data/core0",
      "dataDir":"/var/solr/data/core0/data/",
      "config":"solrconfig.xml",
      "schema":"managed-schema",
      "startTime":"2024-01-15T08:30:00.000Z",
      "uptime":8640000,
      "index":{
        "numDocs":154823,
        "maxDoc":154823,
        "deletedDocs":0,
        "version":42,
        "segmentCount":3,
        "current":true,
        "hasDeletions":false,
        "directory":"org.apache.lucene.store.NRTCachingDirectory",
        "sizeInBytes":98234567,
        "size":"93.68 MB"
      }
    }
  }
}`
}

// SpringBootHealthJSON returns a fake Spring Boot /actuator/health response.
func SpringBootHealthJSON() string {
	return `{"status":"UP","components":{"db":{"status":"UP","details":{"database":"PostgreSQL","validationQuery":"isValid()"}},"diskSpace":{"status":"UP","details":{"total":107374182400,"free":54223118336,"threshold":10485760,"path":"/app/.","exists":true}},"ping":{"status":"UP"}}}`
}

// SpringBootEnvJSON returns a fake Spring Boot /actuator/env response
// with realistic-looking properties.
func SpringBootEnvJSON() string {
	return `{"activeProfiles":["production"],"propertySources":[{"name":"server.ports","properties":{"local.server.port":{"value":"8080"}}},{"name":"systemProperties","properties":{"java.runtime.name":{"value":"OpenJDK Runtime Environment"},"java.vm.version":{"value":"17.0.8+7"},"os.name":{"value":"Linux"}}},{"name":"applicationConfig: [classpath:/application.yml]","properties":{"spring.datasource.url":{"value":"jdbc:postgresql://db.internal:5432/appdb"},"spring.datasource.username":{"value":"app_user"},"spring.datasource.password":{"value":"******"},"spring.redis.host":{"value":"redis.internal"},"server.port":{"value":"8080"},"management.endpoints.web.exposure.include":{"value":"health,env,info"}}}]}`
}

// MOVEitLoginPage returns a fake MOVEit Transfer login page.
func MOVEitLoginPage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>MOVEit Transfer</title>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<style>body{font-family:Segoe UI,Arial,sans-serif;background:#f0f2f5;margin:0}.header{background:#00437b;color:#fff;padding:12px 24px;font-size:1.2em}.login-box{max-width:400px;margin:60px auto;background:#fff;border:1px solid #ddd;border-radius:4px;padding:2em;box-shadow:0 2px 8px rgba(0,0,0,.1)}h2{color:#00437b;margin-top:0}label{display:block;margin-bottom:.3em;color:#555}input[type=text],input[type=password]{width:100%;padding:8px;margin-bottom:1em;border:1px solid #ccc;border-radius:3px;box-sizing:border-box}input[type=submit]{background:#00437b;color:#fff;padding:10px 20px;border:none;border-radius:3px;cursor:pointer}input[type=submit]:hover{background:#002f57}.footer{text-align:center;color:#999;font-size:.8em;margin-top:2em}</style>
</head>
<body>
<div class="header">MOVEit Transfer</div>
<div class="login-box">
<h2>Sign In</h2>
<form action="/human.aspx" method="post">
<label for="user">Username:</label>
<input type="text" name="Username" id="user" autocomplete="username"/>
<label for="pass">Password:</label>
<input type="password" name="Password" id="pass" autocomplete="current-password"/>
<input type="submit" value="Sign In"/>
</form>
<div class="footer">Progress MOVEit Transfer 2023.0.1</div>
</div>
</body>
</html>`
}

// StrutsShowcasePage returns a fake Apache Struts2 showcase page.
func StrutsShowcasePage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>Struts2 Showcase</title>
<style>body{font-family:Arial,sans-serif;margin:0;padding:0}#header{background:#4a7ab5;color:#fff;padding:10px 20px}#header h1{margin:0;font-size:1.3em}#nav{background:#eee;padding:10px 20px;border-bottom:1px solid #ddd}#nav a{margin-right:15px;color:#336;text-decoration:none}#content{padding:20px;max-width:800px}.info{background:#f9f9f9;border:1px solid #ddd;padding:1em;border-radius:3px;margin-top:1em}code{background:#eee;padding:2px 6px;border-radius:2px}</style>
</head>
<body>
<div id="header"><h1>Struts2 Showcase</h1></div>
<div id="nav"><a href="/struts2-showcase/">Home</a><a href="/struts2-showcase/showcase.action">Showcase</a><a href="/struts2-showcase/validation/">Validation</a><a href="/struts2-showcase/ajax/">AJAX</a></div>
<div id="content">
<h2>Welcome to the Struts2 Showcase Application</h2>
<p>This application demonstrates various features of the Apache Struts 2 framework.</p>
<div class="info"><strong>Version:</strong> Apache Struts 2.5.30<br/><strong>Server:</strong> Apache Tomcat/9.0.65</div>
</div>
</body>
</html>`
}

// ConfluenceLoginPage returns a fake Atlassian Confluence login page.
func ConfluenceLoginPage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>Log In - Confluence</title>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<style>body{font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,sans-serif;background:#f4f5f7;margin:0}.logo{text-align:center;padding:40px 0 20px}.logo h1{color:#172b4d;font-size:1.5em}.login-box{max-width:400px;margin:0 auto;background:#fff;border-radius:3px;box-shadow:0 1px 3px rgba(0,0,0,.12);padding:2em}label{display:block;margin-bottom:.3em;color:#6b778c;font-size:.85em;font-weight:600}input[type=text],input[type=password]{width:100%;padding:8px;margin-bottom:1em;border:2px solid #dfe1e6;border-radius:3px;box-sizing:border-box;font-size:1em}input[type=text]:focus,input[type=password]:focus{border-color:#4c9aff;outline:none}input[type=submit]{width:100%;padding:10px;background:#0052cc;color:#fff;border:none;border-radius:3px;font-size:1em;cursor:pointer;font-weight:500}input[type=submit]:hover{background:#0065ff}.footer{text-align:center;margin-top:2em;color:#97a0af;font-size:.8em}</style>
</head>
<body>
<div class="logo"><h1>Confluence</h1></div>
<div class="login-box">
<form action="/dologin.action" method="post">
<label for="username">Username</label>
<input type="text" name="os_username" id="username" autocomplete="username"/>
<label for="password">Password</label>
<input type="password" name="os_password" id="password" autocomplete="current-password"/>
<input type="submit" value="Log in"/>
</form>
</div>
<div class="footer">Atlassian Confluence 8.5.3</div>
</body>
</html>`
}

// ExposedEnvFile returns a fake .env file with realistic-looking
// but non-functional credentials.
func ExposedEnvFile() string {
	return `APP_NAME=MyApp
APP_ENV=production
APP_DEBUG=false
APP_URL=https://app.example.com

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=myapp_prod
DB_USERNAME=myapp_user
DB_PASSWORD=super_secret_password_2024

REDIS_HOST=127.0.0.1
REDIS_PASSWORD=redis_secret_pass

MAIL_MAILER=smtp
MAIL_HOST=smtp.mailgun.org
MAIL_PORT=587
MAIL_USERNAME=postmaster@mg.example.com
MAIL_PASSWORD=mail_key_abc123

AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
AWS_DEFAULT_REGION=us-east-1
AWS_BUCKET=myapp-uploads

STRIPE_KEY=sk_live_fake_stripe_key_do_not_use
STRIPE_SECRET=whsec_fake_webhook_secret
`
}
