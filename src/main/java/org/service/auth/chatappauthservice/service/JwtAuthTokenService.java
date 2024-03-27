package org.service.auth.chatappauthservice.service;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwe;
import io.jsonwebtoken.JwtParser;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.io.Decoders;
import io.jsonwebtoken.security.Keys;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.entity.enums.TokenType;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.util.Date;

@Service
public class JwtAuthTokenService implements AuthTokenService {

	private static final int ACCESS_EXPIRATION = 2 * 60 * 60; // 2 hours

	private static final int REFRESH_EXPIRATION = 7 * 24 * 60 * 60; // 7 days

	@Value("${application.jwt.access_token_secret}")
	private String secretAccessKey;

	@Value("${application.jwt.refresh_token_secret}")
	private String secretRefreshKey;

	public JwtAuthTokenService() {
	}

	@Override
	public String createAccessToken(User user) {
		return createToken(user, secretAccessKey, ACCESS_EXPIRATION);
	}

	@Override
	public String createRefreshToken(User user) {
		return createToken(user, secretRefreshKey, REFRESH_EXPIRATION);
	}

	public String createToken(User user, String secretKey, int expiration) {
		Date createdAt = new Date(System.currentTimeMillis());
		Date expireAt = new Date(createdAt.getTime() + expiration);

		return Jwts.builder()
			.header()
			.type("JWT")
			.and()
			.claims()
			.subject(String.valueOf(user.getUserId()))
			.add("email", user.getEmail())
			.add("username", user.getUsername())
			.add("role", user.getRole())
			.issuedAt(createdAt)
			.expiration(expireAt)
			.and()
			.signWith(getSecretKey(secretKey))
			.compact();
	}

	@Override
	public boolean isTokenExpired(String jwtToken, TokenType type) {
		Claims claims;
		switch (type) {
			case ACCESS_TOKEN -> claims = extractAllClaims(jwtToken, secretAccessKey);
			case REFRESH_TOKEN -> claims = extractAllClaims(jwtToken, secretRefreshKey);
			default -> throw new RuntimeException("Not handle yet");
		}

		return claims.getExpiration().after(new Date());
	}

	private Claims extractAllClaims(String jwtToken, String secretKey) {
		JwtParser parser = Jwts.parser().verifyWith(getSecretKey(secretKey)).build();
		return parser.parse(jwtToken).accept(Jwe.CLAIMS).getPayload();
	}

	private SecretKey getSecretKey(String secretKey) {
		byte[] keyBytes = Decoders.BASE64.decode(secretKey);
		return Keys.hmacShaKeyFor(keyBytes);
	}

}
