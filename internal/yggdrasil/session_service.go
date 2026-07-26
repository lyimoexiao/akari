package yggdrasil

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) JoinServer(ctx context.Context, req *JoinReq, clientIP string) error {
	token, err := s.findValidToken(ctx, req.AccessToken, "")
	if err != nil {
		return err
	}
	if token.ProfileUUID == "" || normalizeUUID(req.SelectedProfile) != token.ProfileUUID {
		return ErrInvalidToken
	}
	session := ServerSession{
		AccessToken: req.AccessToken,
		ProfileUUID: token.ProfileUUID,
		IP:          clientIP,
	}
	if err := s.sessions.Save(ctx, req.ServerID, session, serverSessionTTL); err != nil {
		return fmt.Errorf("cache join session: %w", err)
	}
	return nil
}

func (s *Service) HasJoined(ctx context.Context, username, serverID, ip string) (*ProfileResp, error) {
	session, err := s.sessions.Load(ctx, serverID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	profile, err := s.getProfileByUUID(ctx, session.ProfileUUID)
	if err != nil {
		return nil, ErrProfileNotFound
	}
	if profile.Name != username {
		return nil, ErrInvalidToken
	}
	if ip != "" && session.IP != "" && session.IP != ip {
		return nil, ErrInvalidToken
	}
	response := s.profileWithTextures(profile, true)
	return &response, nil
}

func (s *Service) GetProfile(ctx context.Context, uuid string, unsigned bool) (*ProfileResp, error) {
	profile, err := s.getProfileByUUID(ctx, normalizeUUID(uuid))
	if err != nil {
		return nil, err
	}
	response := s.profileWithTextures(profile, !unsigned)
	return &response, nil
}

func (s *Service) GetProfilesByName(ctx context.Context, names []string) ([]ProfileResp, error) {
	profiles, err := s.getProfilesByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	result := make([]ProfileResp, len(profiles))
	for index := range profiles {
		result[index] = toProfileResp(&profiles[index])
	}
	return result, nil
}

func (s *Service) UserStatus(ctx context.Context, userID uint) (*UserStatusResp, error) {
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profiles, err := s.getProfilesForUser(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	response := &UserStatusResp{}
	if len(profiles) > 0 {
		response.HasProfile = true
		response.ProfileUUID = profiles[0].UUID
		response.ProfileName = profiles[0].Name
	}
	lastToken, err := s.repository.LastLoginToken(ctx, user.Email)
	if err == nil {
		lastLoginAt := lastToken.IssuedAt.Format(time.RFC3339)
		response.LastLoginAt = &lastLoginAt
		response.LastLoginIP = lastToken.LoginIP
	}
	return response, nil
}

func (s *Service) Metadata() *MetadataResp {
	publicKey := ""
	if s.signer != nil {
		publicKey = s.signer.PublicKey()
	}
	return &MetadataResp{
		Meta: map[string]any{
			"serverName":              s.settings.ServerName,
			"implementationName":      s.settings.ImplementationName,
			"implementationVersion":   s.settings.ImplementationVersion,
			"feature.non_email_login": true,
		},
		SkinDomains:        []string{},
		SignaturePublickey: publicKey,
	}
}
