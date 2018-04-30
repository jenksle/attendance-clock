<!doctype html>

<html>
    <head>
        <title>RJW Clocking Application</title>
        <link rel="stylesheet" type="text/css" href="/assets/css/reset.css" />
        <link rel="stylesheet" type="text/css" href="/assets/css/core.css" />
        <link rel="stylesheet" type="text/css" href="/assets/css/main.css" />
        <meta charset="UTF-8">
    </head>

    <body class="stage_1" onload="startTime();">
        <div id="container">

            <div id="sidebar">
 
                <a href="/"> <img src="/assets/img/header.png" id="header" /></a>

                <ul id="stages">

                    <li id="clock">
                        <a class="clock_inner"><div id="clock2">{{.ServerTime}}</div></a>
                    </li>

                    <li id="stage_1" class="active">
                        <a class="stage_inner" href="/" onClick="return false;"> <span class="enum">1</span>
                            <h2 class="question">Pick a<br />department</h2>
                        </a>
                    </li>

                    <li id="stage_2">
                        <a class="stage_inner" href="#" onClick="return false;"> <span class="enum">2</span>
                            <h2 class="question">Choose an<br />employee</h2>
                        </a>
                    </li>

                    <li id="stage_3">
                        <a class="stage_inner" href="#" onClick="return false;"> <span class="enum">3</span>
                            <h2 class="question">Clock in/out,<br />check logs</h2>
                        </a>
                    </li>
                </ul>

            </div>


            <div id="main" class="stage_1">
                {{range .Departments}}
                    <a class="inset" href="/employees/{{.DepartmentID}}">{{.DepartmentName}}</a>
                {{end}}
            </div>
		</div>

        <script src="/assets/js/jquery-3.3.1.min.js"></script>
        <script src="/assets/js/time.js"></script>
        

    </body>
</html>

