package authority

import (
	"fmt"
	"github.com/casbin/casbin/v2"
)

const (
	p                     = "p"
	policy                = "policy"
	g                     = "g"
	g2                    = "g2"
	grouping              = "grouping"
	grouping2             = "grouping2"
	domainMember          = "domain_member"
	departmentInheritance = "department_inheritance"
)

type PolicyOperation struct {
	PolicyType string   // policy/p/grouping/g/grouping2/g2
	Params     []string // 策略参数列表
	CheckExist bool     // 是否检查已存在
}

type AuthManager struct {
	e *casbin.Enforcer
}

func NewAuthManager(e *casbin.Enforcer) *AuthManager {
	return &AuthManager{e: e}
}

// ApplyPolicy 通用策略操作方法
func (m *AuthManager) ApplyPolicy(op PolicyOperation) (exists bool, err error) {
	if op.CheckExist {
		switch op.PolicyType {
		case p, policy:
			exists, err = m.e.HasPolicy(op.Params)
		case g, grouping:
			exists, err = m.e.HasGroupingPolicy(op.Params)
		case g2, grouping2:
			exists, err = m.e.HasNamedGroupingPolicy(g2, op.Params)
		default:
			return false, fmt.Errorf("不支持的策略类型: %s", op.PolicyType)
		}

		return exists, nil
	}

	switch op.PolicyType {
	case p, policy:
		exists, err = m.e.AddPolicy(op.Params)
		return exists, err
	case g, grouping:
		exists, err = m.e.AddGroupingPolicy(op.Params)
		return exists, err
	case g2, grouping2:
		exists, err = m.e.AddNamedGroupingPolicy(g2, op.Params)
		return exists, err
	default:
		return exists, fmt.Errorf("未知的策略类型: %s", op.PolicyType)
	}
}

// SavePolicies 保存策略到数据库
func (m *AuthManager) SavePolicies() error {
	return m.e.SavePolicy()
}

// AddDomainInheritance 添加域继承关系（g2策略）
func (m *AuthManager) AddDomainInheritance(childDomain, parentDomain, relation string) error {
	op := PolicyOperation{
		PolicyType: g2,
		Params:     []string{childDomain, parentDomain, relation},
		CheckExist: true,
	}
	_, err := m.ApplyPolicy(op)
	return err
}

// AddPolicy 添加权限策略（p策略）
func (m *AuthManager) AddPolicy(role, domain, path, action string) error {
	op := PolicyOperation{
		PolicyType: p,
		Params:     []string{role, domain, path, action},
		CheckExist: true,
	}
	_, err := m.ApplyPolicy(op)
	return err
}

// AddUserRole 添加用户角色（g策略）
func (m *AuthManager) AddUserRole(user, role, domain string) error {
	op := PolicyOperation{
		PolicyType: g,
		Params:     []string{user, role, domain},
		CheckExist: true,
	}
	_, err := m.ApplyPolicy(op)
	return err
}

// AddDepartmentInheritance 添加部门继承（g2策略）
func (m *AuthManager) AddDepartmentInheritance(dept, company string) error {
	return m.AddDomainInheritance(dept, company, departmentInheritance)
}

// AddDomainMember 添加域成员（g2策略）
func (m *AuthManager) AddDomainMember(domain string) error {
	op := PolicyOperation{
		PolicyType: g2,
		Params:     []string{domain, domain, domainMember},
		CheckExist: true,
	}
	_, err := m.ApplyPolicy(op)
	return err
}
