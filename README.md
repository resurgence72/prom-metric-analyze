### prom-metric-analyze

> 项目基于 mimirtool 进行封装，实现了对 prometheus 的指标基数优化；
>
> 其能实现根据指定的 grafana + prometheus rule 文件 分析出当前所使用的指标，同时分析 TSDB 侧来得出哪些
> 指标是未被使用的，分析生成优化relabel配置；
>
> 拿到结果后，我们就可以在 metric_relabel/write_relabel 阶段(根据需求)，来对相关无用指标进行 keep，从而实现 prometheus series的精简，
> 优化磁盘大小，优化内存，降低笛卡尔积，查询提速；



### 运维指南

> #### 0. 配置文件说明
>
> 前提，准备好grafana api-key + 当前所用到的 prometheus rule/record 文件；
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
> mimirtool_dir: ./bin     				  # mimirtool 二进制的目录，如果检测到没有，会去github下载二进制到此目录下
> 
> optimization_relabel_type: metric_relabel_configss   # 最后需要生成的优化配置段 metric_relabel_configss/write_relabel_configs
> ```
>
> #### 1. 项目构建 运行
>
> ```sh
> # 构建
> cd prom-metric-analyze/cmd && go build -o prom-metric-analyze main.go
> 
> # 执行
> ./prom-metric-analyze -config.file ./analyze.yaml
> ```
>
> #### 2. 运行效果
>
> ```shell
> root@DESKTOP-269NF9M:/mnt/d/project/prom-metric-analyze/cmd# ./prom-metric-analyze -config.file ./analyze.yaml
> WARN[0000] ./bin/mimirtool mimirtool binary is exist
> WARN[0000] start analyze
> WARN[0000] analyze command: /bin/bash -c ./bin/mimirtool analyze grafana --address http://10.0.0.100:30030 --key eyJrIjoiVjdJdGkzczFERmJ2dTB1WkRZZGRJQ05GeVBDNUt4SmUiLCJuIjoibWltaXJ0b29sIiwiaWQiOjF9 --output .output/metric_in_10.0.0.100_30030_grafana.json
> WARN[0001] analyze command: /bin/bash -c ./bin/mimirtool analyze rule-file ./rules/*.yaml --output .output/rule_in_10.0.0.100_30090_prometheus.json
> WARN[0001] analyze command: /bin/bash -c ./bin/mimirtool analyze prometheus --address http://10.0.0.100:30090 --grafana-metrics-file .output/metric_in_10.0.0.100_30030_grafana.json --ruler-metrics-file .output/rule_in_10.0.0.100_30090_prometheus.json --output .output/.analyze-result.json
> WARN[0003] mimirtool output:time="2023-03-15T21:22:32+08:00" level=info msg="Found 1289 metric names"
> time="2023-03-15T21:22:32+08:00" level=info msg="6904 active series are being used in dashboards"
> time="2023-03-15T21:22:33+08:00" level=info msg="40457 active series are NOT being used in dashboards"
> time="2023-03-15T21:22:33+08:00" level=info msg="243 in use active series metric count"
> time="2023-03-15T21:22:33+08:00" level=info msg="1045 not in use active series metric count"
> 
> 
> WARN[0003] The optimized configuration segment has been generated in the ./metrics_analyze_result directory
> ```
>
> 完成后执行 ls -l 即可在当前目录下看到生成的优化结果目录
>
> ```shell
> root@DESKTOP-269NF9M:/mnt/d/project/prom-metric-analyze/cmd# ls -l
> total 0
> ...
> -rwxrwxrwx 1 root1 root1  311 Mar 15 21:02 analyze.yaml
> drwxrwxrwx 1 root1 root1 4096 Mar 15 21:22 metrics_analyze_result    # 优化结果目录
> drwxrwxrwx 1 root1 root1 4096 Mar 14 23:39 rules
> 
> 
> # 进入 metrics_analyze_result 可看到当前根据 metric 前缀分别生成了相对应的 xxx_relabel_configs 配置
> root@DESKTOP-269NF9M:/mnt/d/project/prom-metric-analyze/cmd/metrics_analyze_result# ll
> total 16
> -rwxrwxrwx 1 root1 root1  151 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_aggregator
> -rwxrwxrwx 1 root1 root1  450 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_alertmanager
> -rwxrwxrwx 1 root1 root1  247 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_apiserver
> -rwxrwxrwx 1 root1 root1  389 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_cluster
> -rwxrwxrwx 1 root1 root1  122 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_code
> -rwxrwxrwx 1 root1 root1  633 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_container
> -rwxrwxrwx 1 root1 root1  302 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_coredns
> -rwxrwxrwx 1 root1 root1   91 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_go
> -rwxrwxrwx 1 root1 root1  583 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_instance
> -rwxrwxrwx 1 root1 root1 1369 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_kube
> -rwxrwxrwx 1 root1 root1  659 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_kubelet
> -rwxrwxrwx 1 root1 root1  285 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_kubeproxy
> -rwxrwxrwx 1 root1 root1   99 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_kubernetes
> -rwxrwxrwx 1 root1 root1  345 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_namespace
> -rwxrwxrwx 1 root1 root1 1518 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_node
> -rwxrwxrwx 1 root1 root1  123 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_others
> -rwxrwxrwx 1 root1 root1  160 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_process
> -rwxrwxrwx 1 root1 root1 2482 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_prometheus
> -rwxrwxrwx 1 root1 root1  148 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_rest
> -rwxrwxrwx 1 root1 root1  186 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_scheduler
> -rwxrwxrwx 1 root1 root1  191 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_storage
> -rwxrwxrwx 1 root1 root1  106 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_volume
> -rwxrwxrwx 1 root1 root1  154 Mar 15 21:22 suggest_for_metric_relabel_configs_with_prefix_workqueue
> ```
>
> 随便打开一个文件看效果
>
> ```yaml
> # root@DESKTOP-269NF9M:cmd/metrics_analyze_result# cat suggest_for_metric_relabel_configs_with_prefix_kubelet
> metric_relabel_configs:
> - action: keep
>   source_labels: [ __name__ ]
>   regex: kubelet_runtime_operations_duration_seconds_bucket|kubelet_pod_worker_duration_seconds_bucket|kubelet_pleg_relist_duration_seconds_bucket|kubelet_cgroup_manager_duration_seconds_bucket|kubelet_runtime_operations_total|kubelet_pod_start_duration_seconds_bucket|kubelet_pleg_relist_interval_seconds_bucket|kubelet_running_containers|kubelet_runtime_operations_errors_total|kubelet_pod_worker_duration_seconds_count|kubelet_cgroup_manager_duration_seconds_count|kubelet_node_name|kubelet_pod_start_duration_seconds_count|kubelet_running_pods|kubelet_pleg_relist_duration_seconds_count
> ```
>
> 可以看到，已经生成了针对 kubelet 前缀的 metric_relabel_configs keep规则，可以直接配置到 prometheus 指定配置段，从而实现精简并优化metric的效果

