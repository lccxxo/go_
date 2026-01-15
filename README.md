# Go 学习项目

这是一个 Go 语言学习项目，包含设计模式、算法实现以及各种实用示例。

## 项目结构

```
go_/
├── design_patterns/      # 设计模式示例
│   ├── simple_factory/   # 简单工厂模式
│   ├── factory_method/   # 工厂方法模式
│   ├── singleton/        # 单例模式
│   ├── observer/         # 观察者模式
│   ├── strategy/         # 策略模式
│   ├── decorator/        # 装饰器模式
│   ├── adapter/          # 适配器模式
│   ├── builder/          # 建造者模式
│   ├── facede/           # 外观模式
│   └── README.md         # 设计模式学习指南
│
├── algorithms/           # 算法实现
│   ├── sorting/          # 排序算法
│   ├── search/           # 查找算法
│   ├── dynamic_programming/  # 动态规划
│   ├── greedy/           # 贪心算法
│   ├── graph/            # 图算法
│   └── README.md         # 算法学习指南
│
├── priorityQueue/        # 优先级队列实现
├── oauth/               # OAuth 认证示例
├── uploadBigFile/       # 大文件上传示例
└── receiveBigFile/      # 大文件接收示例
```

## 快速开始

### 运行设计模式示例

```bash
# 单例模式
cd design_patterns/singleton
go run main.go

# 观察者模式
cd design_patterns/observer
go run main.go

# 策略模式
cd design_patterns/strategy
go run main.go
```

### 运行算法示例

```bash
# 排序算法
cd algorithms/sorting
go run main.go

# 查找算法
cd algorithms/search
go run main.go

# 动态规划
cd algorithms/dynamic_programming
go run main.go
```

## 学习资源

- [设计模式学习指南](./design_patterns/README.md)
- [算法学习指南](./algorithms/README.md)

## 设计模式

本项目包含以下设计模式：

### 创建型模式
- ✅ 简单工厂模式
- ✅ 工厂方法模式
- ✅ 单例模式
- ✅ 建造者模式

### 结构型模式
- ✅ 适配器模式
- ✅ 装饰器模式
- ✅ 外观模式

### 行为型模式
- ✅ 观察者模式
- ✅ 策略模式

## 算法实现

本项目包含以下算法：

- ✅ **排序算法**: 冒泡、快速、归并、插入、选择排序
- ✅ **查找算法**: 线性查找、二分查找
- ✅ **动态规划**: 斐波那契、爬楼梯、LCS、0-1背包
- ✅ **贪心算法**: 活动选择、找零、分数背包
- ✅ **图算法**: DFS、BFS、拓扑排序、Dijkstra

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License