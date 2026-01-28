package services

// ValidateCredentials 验证飞书凭证是否有效
func (s *LarkService) ValidateCredentials() error {
	// 尝试获取访问令牌来验证凭证
	_, err := s.GetTenantAccessToken()
	return err
}