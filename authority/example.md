# Casbin 权限管理系统使用指南

### 文档包含：
   1. 基础配置方法
   2. 主要功能场景演示
   3. 不同策略类型的组合使用
   4. 实际业务中的典型应用案例

## 快速开始
```go
    package main
    
    import (
        "pkgs/authority"
        "github.com/casbin/casbin/v2"
        pgadapter "github.com/casbin/casbin-pg-adapter"
    )
    
    func main() {
        // 初始化enforcer和适配器
        modelConfigPath := "conf/rbac_model.conf"
        policyCSVPath := "policy.csv"
        enforcer, _ := casbin.NewEnforcer(modelConfigPath, policyCSVPath)
        
        authMgr := auth_manager.NewAuthManager(enforcer)
    }
```

## 基本用法
### 1. 设置公司组织结构
```go
    // 创建总公司与系统域的继承关系
    authMgr.AddDomainInheritance(
        "总公司", 
        "系统域",
        "隶属关系",
    )
    
    // 建立分公司与总公司的继承关系
    authMgr.AddDomainInheritance(
        "北京分公司",
        "总公司",
        "分支关系",
    )
    
    // 在分公司下创建研发部门
    authMgr.AddDepartmentInheritance(
        "研发部",
        "北京分公司",
    )
```

### 2. 设置用户权限
```go
    // 为张伟分配管理员角色
    authMgr.AddUserRole(
        "张伟",
        "分公司管理员", 
        "北京分公司",
    )
    
    // 为分公司管理员分配权限
    authMgr.AddPolicy(
        "分公司管理员",
        "北京分公司",
        "/*",
        "全权限",
    )
    
    // 为李娜分配员工角色
    authMgr.AddUserRole(
        "李娜",
        "普通员工",
        "研发部",
    )
    
    // 为研发部员工设置文档权限
    authMgr.AddPolicy(
        "普通员工",
        "研发部",
        "/文档/*",
        "只读",
    )
```

### 3. 跨域权限继承
```go
    // 总公司可查看所有分支机构的报表
    authMgr.AddPolicy(
        "总公司管理员", 
        "总公司",
        "/分支机构/*/报表", 
        "查看",
    )
    
    // 域继承使分公司自动获得权限
    authMgr.AddDomainInheritance(
        "北京分公司",
         "总公司",
         "数据继承",
    )
```
### 4. 临时权限授予

```go
    // 给王强临时项目权限（不检查是否已存在）
    authMgr.ApplyPolicy(auth_manager.PolicyOperation{
        PolicyType: "p",
        Params:     []string{"临时成员", "研发项目A", "/项目文档", "编辑"},
        CheckExist: false,
    })
```

## 注意事项
1. 参数顺序必须严格匹配模型定义
2. 重要操作建议设置 CheckExist: true 避免重复
3. 修改后务必调用 SavePolicies() 保存到数据库 
4. 域关系建议使用英文标识（如 company:A）