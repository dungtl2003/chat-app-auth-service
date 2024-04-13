package org.service.auth.chatappauthservice.service;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.ExpiredJwtException;
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

	private static final long ACCESS_EXPIRATION = 2 * 60 * 60 * 1000; // 2 hours

	private static final long REFRESH_EXPIRATION = 7 * 24 * 60 * 60 * 1000; // 7 days

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

	private String createToken(User user, String secretKey, long expiration) {
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
		try {
			String secretKey;
			switch (type) {
				case ACCESS_TOKEN -> secretKey = secretAccessKey;
				case REFRESH_TOKEN -> secretKey = secretRefreshKey;
				default -> throw new RuntimeException("Invalid token type");
			}
			JwtParser parser = getParser(secretKey);
			parser.parseSignedClaims(jwtToken);
		}
		catch (ExpiredJwtException eje) {
			return true;
		}

		return false;
	}

	private Claims extractAllClaims(String jwtToken, String secretKey) throws ExpiredJwtException {
		JwtParser parser = getParser(secretKey);
		return parser.parseSignedClaims(jwtToken).getPayload();
	}

	private SecretKey getSecretKey(String secretKey) {
		byte[] keyBytes = Decoders.BASE64.decode(secretKey);
		return Keys.hmacShaKeyFor(keyBytes);
	}

	private JwtParser getParser(String secretKey) {
		return Jwts.parser().verifyWith(getSecretKey(secretKey)).build();
	}

}
