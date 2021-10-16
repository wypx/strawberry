{{ define "strawberry_mobile2" }}
<!DOCTYPE html>
<html class="no-js css-menubar" lang="zh">
  <head>
    {{template "strawberry_header" .}}
  </head>
  <body class="animsition app-travel">
    {{template "strawberry_top" .}}
    {{template "strawberry_left" .}}
    {{template "strawberry_grid" .}}
      
    <!-- Page -->
    <div id="strawberry_app">


    </div>

    {{template "strawberry_footer" .}}

  </body>
</html>
{{ end }}