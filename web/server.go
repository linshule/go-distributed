package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/linshule/go-distributed/registry"
)

type webServer struct{}

var webPage = `
<!DOCTYPE html>
<html>
<head>
    <title>分布式服务管理</title>
    <meta charset="utf-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .section {
            margin-bottom: 30px;
        }
        h2 {
            color: #555;
            border-bottom: 2px solid #007bff;
            padding-bottom: 10px;
        }
        button {
            background-color: #007bff;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 4px;
            cursor: pointer;
            margin: 5px;
        }
        button:hover {
            background-color: #0056b3;
        }
        button.danger {
            background-color: #dc3545;
        }
        button.danger:hover {
            background-color: #c82333;
        }
        #services {
            margin-top: 20px;
        }
        .service-card {
            background: #f8f9fa;
            border-left: 4px solid #007bff;
            padding: 15px;
            margin: 10px 0;
            border-radius: 4px;
        }
        .service-name {
            font-weight: bold;
            font-size: 18px;
            color: #333;
        }
        .service-url {
            color: #666;
            margin-top: 5px;
        }
        .status {
            display: inline-block;
            padding: 3px 8px;
            border-radius: 3px;
            font-size: 12px;
            margin-left: 10px;
        }
        .status.online {
            background-color: #28a745;
            color: white;
        }
        #log-form {
            margin-top: 20px;
        }
        input, textarea {
            width: 100%;
            padding: 10px;
            margin: 5px 0;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        #log-result {
            margin-top: 10px;
            padding: 10px;
            border-radius: 4px;
            display: none;
        }
        .success {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
        }
        .error {
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }
        .refresh-info {
            color: #666;
            font-size: 12px;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>分布式服务管理系统</h1>

        <div class="section">
            <h2>服务列表</h2>
            <button onclick="refreshServices()">刷新服务列表</button>
            <div id="services"></div>
            <div class="refresh-info">点击"刷新服务列表"查看已注册的服务</div>
        </div>

        <div class="section">
            <h2>发送日志</h2>
            <div id="log-form">
                <input type="text" id="log-message" placeholder="输入日志消息...">
                <button onclick="sendLog()">发送日志</button>
                <div id="log-result"></div>
            </div>
        </div>

        <div class="section">
            <h2>图书馆服务</h2>
            <button onclick="listBooks()">查看图书</button>
            <button onclick="listBorrowRecords()">查看借阅记录</button>
            <div id="library-result"></div>
        </div>
    </div>

    <script>
        async function refreshServices() {
            try {
                const response = await fetch('/services');
                const services = await response.json();
                const container = document.getElementById('services');
                if (services.length === 0) {
                    container.innerHTML = '<p>暂无注册服务</p>';
                } else {
                    container.innerHTML = services.map(s =>
                        '<div class="service-card">' +
                            '<span class="service-name">' + s.serviceName + '</span>' +
                            '<span class="status online">在线</span>' +
                            '<div class="service-url">' + s.serviceUrl + '</div>' +
                        '</div>'
                    ).join('');
                }
            } catch (error) {
                console.error('Error:', error);
                document.getElementById('services').innerHTML = '<p class="error">获取服务列表失败</p>';
            }
        }

        async function sendLog() {
            const message = document.getElementById('log-message').value;
            if (!message) {
                showResult('log-result', '请输入日志消息', 'error');
                return;
            }
            try {
                const response = await fetch('http://localhost:4000/log', {
                    method: 'POST',
                    body: message
                });
                if (response.ok) {
                    showResult('log-result', '日志发送成功！', 'success');
                    document.getElementById('log-message').value = '';
                } else {
                    showResult('log-result', '日志发送失败', 'error');
                }
            } catch (error) {
                showResult('log-result', '日志服务未启动或无法访问', 'error');
            }
        }

        async function listBooks() {
            try {
                const response = await fetch('http://localhost:5000/library/books');
                const books = await response.json();
                const container = document.getElementById('library-result');
                container.innerHTML = '<h3>图书列表</h3>' +
                    '<pre>' + JSON.stringify(books, null, 2) + '</pre>';
            } catch (error) {
                document.getElementById('library-result').innerHTML = '<p class="error">图书馆服务未启动或无法访问</p>';
            }
        }

        async function listBorrowRecords() {
            try {
                const response = await fetch('http://localhost:5000/library/borrow');
                const records = await response.json();
                const container = document.getElementById('library-result');
                container.innerHTML = '<h3>借阅记录</h3>' +
                    '<pre>' + JSON.stringify(records, null, 2) + '</pre>';
            } catch (error) {
                document.getElementById('library-result').innerHTML = '<p class="error">图书馆服务未启动或无法访问</p>';
            }
        }

        function showResult(elementId, message, className) {
            const el = document.getElementById(elementId);
            el.textContent = message;
            el.className = className;
            el.style.display = 'block';
        }

        // 页面加载时自动刷新服务列表
        window.onload = refreshServices;
    </script>
</body>
</html>
`

type web struct{}

func (w webServer) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/", "/web":
		// 返回Web页面
		t, err := template.New("web").Parse(webPage)
		if err != nil {
			log.Println(err)
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
		t.Execute(wr, nil)
	case "/services":
		// 代理到注册中心获取服务列表
		regs, err := registry.GetServices()
		if err != nil {
			log.Println(err)
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
		wr.Header().Set("Content-Type", "application/json")
		json.NewEncoder(wr).Encode(regs)
	default:
		http.NotFound(wr, r)
	}
}

// RegisterHandlers 注册HTTP处理器
func RegisterHandlers() {
	http.Handle("/web", &webServer{})
}
