# gosifter

RESTful API 的基本思想，是同一个资源实体在不同的场景下存在不同的 representation（表示视图）。  

从权限管理的角度，同一个资源可能存在如下的访问权限级别：

- L0: 可公开
- L1: 保密
- L2: 高度保密
- L3: 绝密

从系统的角度，访问者角色可能包括：

- 普通用户
- 开发者（可能存在分级）
- 内部服务（可能存在分级）

通过将「访问者角色」映射至「可访问的资源权限级别」，可以进行数据分级以及后续的数据脱敏等处理。