package authmgr

import "html/template"

// templates

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.New(tmIndex).Parse(index))
	template.Must(tmpl.New(tmCallback).Parse(success))
}

const (
	index = `
	<html>

	<head>
		<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
		<link rel="stylesheet" type="text/css"
			  href="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.5/css/materialize.min.css">
		<style>
			body {
				display: flex;
				min-height: 100vh;
				flex-direction: column;
			}
			main {
				flex: 1 0 auto;
			}
			body {
				background: #fff;
			}
			.input-field input[type=date]:focus + label,
			.input-field input[type=text]:focus + label,
			.input-field input[type=email]:focus + label,
			.input-field input[type=password]:focus + label {
				color: #e91e63;
			}
			.input-field input[type=date]:focus,
			.input-field input[type=text]:focus,
			.input-field input[type=email]:focus,
			.input-field input[type=password]:focus {
				border-bottom: 2px solid #e91e63;
				box-shadow: none;
			}
		</style>
	</head>
	
	<body>
	<div class="section"></div>
	<main>
		<center>
			<div class="section"></div>
			<div class="container">
				<div class="z-depth-1 grey lighten-4 row"
					 style="display: inline-block; padding: 32px 48px 0px 48px; border: 1px solid #EEE;">
	
					<div class='row'>
						<div class='col s12'>
						Please login with Google to allow <b>{{.AppName}}</b> to access Google Sheets.<br/>
						This is done only once.
						</div>
					</div>
	
					<div class="row">
						<div class="btn white darken-6 col s12">
							<a href="/login" style="text-transform:none">
								<div class="left">
									<img width="30px" alt="Google &quot;G&quot; Logo"
										 src="https://upload.wikimedia.org/wikipedia/commons/thumb/5/53/Google_%22G%22_Logo.svg/512px-Google_%22G%22_Logo.svg.png"/>
								</div>
								Login with Google
							</a>
						</div>
					</div>
				</div>
			</div>
		</center>
	
		<div class="section"></div>
		<div class="section"></div>
	</main>
	
	<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.2.1/jquery.min.js"></script>
	<script type="text/javascript"
			src="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.5/js/materialize.min.js"></script>
	</body>
	
	</html>
`

	success = `
<html>

<head>
	<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
	<link rel="stylesheet" type="text/css"
		  href="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.5/css/materialize.min.css">
	<style>
		body {
			display: flex;
			min-height: 100vh;
			flex-direction: column;
		}
		main {
			flex: 1 0 auto;
		}
		body {
			background: #fff;
		}
		.input-field input[type=date]:focus + label,
		.input-field input[type=text]:focus + label,
		.input-field input[type=email]:focus + label,
		.input-field input[type=password]:focus + label {
			color: #e91e63;
		}
		.input-field input[type=date]:focus,
		.input-field input[type=text]:focus,
		.input-field input[type=email]:focus,
		.input-field input[type=password]:focus {
			border-bottom: 2px solid #e91e63;
			box-shadow: none;
		}
	</style>
</head>

<body>
<div class="section"></div>
<main>
	<center>
		<div class="section"></div>
		<div class="container">
			<div class="z-depth-1 grey lighten-4 row"
				 style="display: inline-block; padding: 32px 48px 0px 48px; border: 1px solid #EEE;">

				<div class='row'>
					<div class='col s12'>
					<img src="data:image/png;base64,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" alt="" />
					Success:  <b>{{.AppName}}</b> was authorised, this tab can be closed.
					</div>
				</div>

			</div>
		</div>
	</center>

	<div class="section"></div>
	<div class="section"></div>
</main>

<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.2.1/jquery.min.js"></script>
<script type="text/javascript"
		src="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.5/js/materialize.min.js"></script>
</body>

</html>
`
)
