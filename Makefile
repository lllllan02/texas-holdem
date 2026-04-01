.PHONY: all backend frontend clean

# 默认目标：同时启动后端和前端
all: 
	@echo "Starting backend and frontend..."
	@make -j2 backend frontend

# 启动 Go 后端服务 (启用 CGO 以支持 sqlite3)
backend:
	@echo "Starting Go backend server on port 8080..."
	@go run main.go

# 启动 Vite 前端服务
frontend:
	@echo "Starting Vite frontend server..."
	@cd frontend && npm run dev

# 清理构建产物和数据库文件
clean:
	@echo "Cleaning up..."
	@rm -f texas.db
	@rm -rf frontend/dist
	@rm -rf frontend/node_modules
	@echo "Clean complete."

# 重新安装前端依赖
install:
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install
	@echo "Installing backend dependencies..."
	@go mod tidy
