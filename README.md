# HarborFlow Go 标注题根仓库

本公开仓库承载 HarborFlow 跨境口岸联合压力测试运营后端的 Go 标注题。

每道题使用唯一的 `tasks/<task_key>/red` 与 `tasks/<task_key>/green` 分支。题目的任务专用测试只存在于 red 基线和 Gomark 私有 intake 材料中；模型执行阶段从对应 green G1 基线开始。

项目主题来源：中新网《皇岗口岸举行首次港深联合压力测试 旅客1分钟内可通过合作查验通道》。
