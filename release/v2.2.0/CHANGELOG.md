# v2.2.0 更新日志

## 数据层优化

- 新增内存快照机制，一次性预加载人员可用时段及历史排班数据
- 消除循环查询带来的性能瓶颈

## 算法升级

- 引入 GRASP 启发式排班算法
- 通过受限候选列表随机构造多组候选解
- 结合空档回溯修复与全局择优，替代原有固定贪心逻辑

## 效果

- 避免顺序依赖导致的人员负担不均
- 在有限迭代内生成分布更均衡、历史负担更轻的排班方案
- 生成速度大幅提升：从秒级提升至毫秒级

---

## 下载

| 平台 | 文件名 |
|------|--------|
| Linux AMD64 | schedule-system-v2.2.0-linux-amd64.zip |
| Linux ARM64 | schedule-system-v2.2.0-linux-arm64.zip |
| Windows AMD64 | schedule-system-v2.2.0-windows-amd64.zip |
| Windows ARM64 | schedule-system-v2.2.0-windows-arm64.zip |
| 前端（电脑端） | schedule-system-v2.2.0-frontend.zip |
| 前端（移动端） | schedule-system-v2.2.0-frontend-mobile.zip |
