package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import io.jsonwebtoken.*;
import io.jsonwebtoken.io.Decoders;
import io.jsonwebtoken.security.Keys;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.configurations.AppConfiguration;
import org.service.auth.chatappauthservice.constants.ErrorMessage;
import org.service.auth.chatappauthservice.entity.enums.TokenState;
import org.service.auth.chatappauthservice.entity.enums.TokenType;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.Date;
import java.util.Map;
import java.util.function.Function;

@Service
public class JwtAuthTokenService implements AuthTokenService {

	private static final Logger logger = LogManager.getLogger(JwtAuthTokenService.class);

	private final AppConfiguration configuration;

	public JwtAuthTokenService(AppConfiguration configuration) {
		this.configuration = configuration;
	}

	@Override
	public String createAccessToken(UserDTO user) {
		return createToken(user, configuration.getSecretAT(), configuration.getATLifespanInMs());
	}

	@Override
	public String createRefreshToken(UserDTO user) {
		return createToken(user, configuration.getSecretRT(), configuration.getRTLifespanInMs());
	}

	@Override
	public String createAccessToken(UserDTO user, long expiration) {
		return createToken(user, configuration.getSecretAT(), expiration);
	}

	@Override
	public String createRefreshToken(UserDTO user, long expiration) {
		return createToken(user, configuration.getSecretRT(), expiration);
	}

	private String createToken(UserDTO user, String secretKey, long expiration) {
		Date createdAt = new Date(System.currentTimeMillis());
		Date expireAt = new Date(createdAt.getTime() + expiration);

		return Jwts.builder()
			.header()
			.type("JWT")
			.and()
			.claims()
			.subject(String.valueOf(user.userId()))
			.add("email", user.email())
			.add("username", user.username())
			.add("role", user.role().getName())
			.issuedAt(createdAt)
			.expiration(expireAt)
			.and()
			.signWith(getSecretKey(secretKey))
			.compact();
	}

	@Override
	public TokenState checkTokenState(String jwtToken, TokenType type) {
		try {
			String secretKey;
			switch (type) {
				case ACCESS_TOKEN -> secretKey = configuration.getSecretAT();
				case REFRESH_TOKEN -> secretKey = configuration.getSecretRT();
				default -> throw new RuntimeException(ErrorMessage.INVALID_TOKEN_TYPE);
			}
			JwtParser parser = getParser(secretKey);
			parser.parseSignedClaims(jwtToken);
		}
		catch (ExpiredJwtException eje) {
			logger.debug("Token is expired");
			return TokenState.EXPIRED;
		}
		catch (JwtException je) {
			logger.debug("Token is invalid");
			return TokenState.INVALID;
		}
		catch (Exception e) {
			throw new RuntimeException(ErrorMessage.UNHANDLED_EXCEPTION);
		}

		return TokenState.VALID;
	}

	private Claims extractAllClaims(String jwtToken, JwtParser parser) throws ExpiredJwtException {
		return parser.parseSignedClaims(jwtToken).getPayload();
	}

	@Override
	public <T> T extractClaim(String jwtToken, Function<Claims, T> claimsResolver, TokenType type) throws JwtException {
		String secretKey;
		switch (type) {
			case ACCESS_TOKEN -> secretKey = configuration.getSecretAT();
			case REFRESH_TOKEN -> secretKey = configuration.getSecretRT();
			default -> throw new RuntimeException("Invalid token type");
		}
		JwtParser parser = getParser(secretKey);
		Claims claims = extractAllClaims(jwtToken, parser);
		return claimsResolver.apply(claims);
	}

	@Override
	@SuppressWarnings("unchecked")
	public Map<String, String> parseBody(String token) throws JsonProcessingException {
		String json = new String(Base64.getDecoder().decode(token.split("\\.")[1]), StandardCharsets.UTF_8);
		ObjectMapper mapper = JsonMapper.builder().findAndAddModules().build();
		return mapper.readValue(json, Map.class);
	}

	private SecretKey getSecretKey(String secretKey) {
		byte[] keyBytes = Decoders.BASE64.decode(secretKey);
		return Keys.hmacShaKeyFor(keyBytes);
	}

	private JwtParser getParser(String secretKey) {
		return Jwts.parser().verifyWith(getSecretKey(secretKey)).build();
	}

}
