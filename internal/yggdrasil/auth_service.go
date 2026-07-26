package yggdrasil

import (
	"context"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/pkg/bcrypt"
)

func (s *Service) Authenticate(ctx context.Context, req *AuthenticateReq, loginIP string) (*AuthenticateResp, error) {
	user, err := s.findUserByLogin(ctx, req.Username)
	if err != nil || !bcrypt.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}
	if s.settings.EmailVerificationEnabled && !user.IsEmailVerified() {
		return nil, ErrEmailNotVerified
	}
	profile, err := s.ensureProfile(ctx, user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("ensure profile: %w", err)
	}
	clientToken := req.ClientToken
	if clientToken == "" {
		clientToken = generateAccessToken()
	}
	profiles, err := s.getProfilesForUser(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("load profiles: %w", err)
	}
	profileUUID := ""
	if len(profiles) == 1 {
		profileUUID = profile.UUID
	}
	token, err := s.createToken(ctx, user.Email, clientToken, profileUUID, loginIP)
	if err != nil {
		return nil, err
	}
	available := make([]ProfileResp, len(profiles))
	for index := range profiles {
		available[index] = toProfileResp(&profiles[index])
	}
	response := &AuthenticateResp{
		AccessToken:       token.AccessToken,
		ClientToken:       token.ClientToken,
		AvailableProfiles: available,
	}
	if profileUUID != "" {
		selected := toProfileResp(profile)
		response.SelectedProfile = &selected
	}
	if req.RequestUser {
		response.User = &YggdrasilUserResp{ID: userUUID(user.Email)}
	}
	return response, nil
}

func (s *Service) Refresh(ctx context.Context, req *RefreshReq) (*RefreshResp, error) {
	oldToken, err := s.repository.FindToken(ctx, req.AccessToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if req.ClientToken != "" && oldToken.ClientToken != req.ClientToken {
		return nil, ErrInvalidToken
	}
	if oldToken.Status == TokenStatusInvalid {
		return nil, ErrInvalidToken
	}
	if time.Now().After(oldToken.ExpiresAt) {
		_ = s.repository.RevokeToken(ctx, oldToken.AccessToken)
		return nil, ErrTokenExpired
	}
	profileUUID := oldToken.ProfileUUID
	if req.SelectedProfile != nil {
		rawUUID := normalizeUUID(req.SelectedProfile.ID)
		profile, profileErr := s.getProfileByUUID(ctx, rawUUID)
		if profileErr != nil || profile.UserEmail != oldToken.UserEmail {
			return nil, ErrInvalidToken
		}
		profileUUID = rawUUID
	}
	if err := s.revokeToken(ctx, &oldToken); err != nil {
		return nil, fmt.Errorf("revoke token: %w", err)
	}
	newToken, err := s.createToken(ctx, oldToken.UserEmail, oldToken.ClientToken, profileUUID, "")
	if err != nil {
		return nil, err
	}
	response := &RefreshResp{
		AccessToken: newToken.AccessToken,
		ClientToken: newToken.ClientToken,
	}
	if profileUUID != "" {
		profile, profileErr := s.getProfileByUUID(ctx, profileUUID)
		if profileErr == nil {
			selected := toProfileResp(profile)
			response.SelectedProfile = &selected
		}
	}
	if req.RequestUser {
		response.User = &YggdrasilUserResp{ID: userUUID(oldToken.UserEmail)}
	}
	return response, nil
}

func (s *Service) Validate(ctx context.Context, req *ValidateReq) error {
	_, err := s.findValidToken(ctx, req.AccessToken, req.ClientToken)
	return err
}

func (s *Service) Invalidate(ctx context.Context, req *InvalidateReq) error {
	token, err := s.repository.FindToken(ctx, req.AccessToken)
	if err != nil {
		return nil
	}
	return s.revokeToken(ctx, &token)
}

func (s *Service) Signout(ctx context.Context, req *SignoutReq) error {
	user, err := s.findUserByLogin(ctx, req.Username)
	if err != nil || !bcrypt.CheckPassword(req.Password, user.Password) {
		return nil
	}
	return s.repository.RevokeValidTokens(ctx, user.Email)
}

func normalizeUUID(uuid string) string {
	if len(uuid) != 32 {
		return uuid
	}
	return uuid[0:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:32]
}
