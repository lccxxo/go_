### 修订历史


<span style="color:red; font-size: 13px; font-weight: bold">遗留问题的优先级按顺序排列，优先级高的优先解决</span>


#### 2024.08.23
- 解决问题：
  - 使用task模块加入cpu检测指标功能
  - 添加告警模块，在monitor模块达到阈值时会处罚警报
- 遗留问题：
  - 需要一个定义阈值的包或者变量
  - 达到阈值时没有相应的处理，直接调用告警模块（需结合实际情况定制告警方法）
  - memory、disk等模块的指标获取
  - notifer模块的告警以及信息发送


#### 2024.08.21
- 解决问题：
    - 定义全局logger实例
    - 添加了docs模块以及历史修订记录
    - 新增task模块，可添加、暂停、回复、删除任务
    - 新增mysql模块，将task信息写入mysql
- 遗留问题：
    - cpu模块未加入到task模块的使用
    - memory、disk等模块的指标获取
    - notifer模块的告警以及信息发送


#### 2024.08.20
- 解决问题：
    - cpu指标的获取
    - disk、task等模块的初始化
    - 添加了linux和windows系统获取cpu使用率的功能;添加docs模块以及历史修订记录
    - 修复了 internal/monitor/cpu_linux.go 在linux系统中获取cpu使用率产生过大的指标的问题
- 遗留问题：
    - log需要调用logger.GetLogger()来获取logger实例，需要定义一个全局的logger实例
    - memory、disk等模块的指标获取
    - notifer模块的告警以及信息发送


#### 2024.08.19
- 解决问题：
    - 初始化了项目，修订了大致的框架
    - 添加config、internal、startup等基础的包
    - 根据.env的APP_ENV参数来选择使用哪种生产环境
    - 添加了cpu、memory、notifer等模块
- 遗留问题：
    - cpu、memory等模块的指标获取
    - notifer模块的发送
    - log需要调用logger.GetLogger()来获取logger实例，需要定义一个全局的logger实例
    - 无disk等模块，后续添加
    - 无task模块，后续添加
