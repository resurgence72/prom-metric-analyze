### prom-metric-analyze

> 项目基于 mimirtool 工具，上层封装实现了对 prometheus 的指标基数优化；
>
> 其能实现根据指定的 grafana + prometheus rule 文件 分析出当前所使用的指标，同时分析 TSDB 侧来得出哪些
> 指标是未被使用的；
>
> 拿到结果后，我们就可以在 metric_relabel/write_relabel 阶段(根据需求)，来对相关指标进行 drop/keep，从而实现 prometheus series的精简，
> 优化磁盘大小，优化内存，降低笛卡尔积，查询提速





#### 运维指南

> 配置文件说明
>
> ```yaml
> grafana:
>   remote_url: http://10.0.0.100:30030   # 需要分析的grafana的http地址
>   api_token: eyJrIjoiVjdJdGkzczFERmJ2dTB1WkRZZGRJQ05GeVBDNUt4SmUiLCJuIjoibWltaXJ0b29sIiwiaWQiOjF9  # 需要申请grafana api-key
> 
> prometheus:
>   remote_url: http://10.0.0.100:30090   # 需要分析的prometheus地址
>   local_rule_file: ./rules/*.yaml       # 需要分析的 prometheus rule 文件，包括rule/record, 支持通配符
> 
> mimirtool_dir: ./bin     				# mimirtool 二进制的目录，如果检测到没有，会去github下载二进制到此目录下
> ```