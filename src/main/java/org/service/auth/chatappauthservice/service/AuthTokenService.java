package org.service.auth.chatappauthservice.service;

import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.entity.enums.TokenType;

public interface AuthTokenService {

	public String createAccessToken(User user);

	public String createRefreshToken(User user);

	public boolean isTokenExpired(String token, TokenType type);

}
