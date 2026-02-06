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
