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

	private static final long DEFAULT_ACCESS_EXPIRATION = Long.parseLong(System.getenv("ACCESS_JWT_LIFESPAN_MS"));

	private static final long DEFAULT_REFRESH_EXPIRATION = Long.parseLong(System.getenv("REFRESH_JWT_LIFESPAN_MS"));

	private static final String SECRET_ACCESS_KEY = System.getenv("ACCESS_JWT_SECRET");

	private static final String SECRET_REFRESH_KEY = System.getenv("REFRESH_JWT_SECRET");

	public JwtAuthTokenService() {
	}

	@Override
	public String createAccessToken(UserDTO user) {
		return createToken(user, SECRET_ACCESS_KEY, DEFAULT_ACCESS_EXPIRATION);
	}

	@Override
	public String createRefreshToken(UserDTO user) {
		return createToken(user, SECRET_REFRESH_KEY, DEFAULT_REFRESH_EXPIRATION);
	}

	@Override
	public String createAccessToken(UserDTO user, long expiration) {
		return createToken(user, SECRET_ACCESS_KEY, expiration);
	}

	@Override
	public String createRefreshToken(UserDTO user, long expiration) {
		return createToken(user, SECRET_REFRESH_KEY, expiration);
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
                case ACCESS_TOKEN -> secretKey = SECRET_ACCESS_KEY;
                case REFRESH_TOKEN -> secretKey = SECRET_REFRESH_KEY;
                default -> throw new RuntimeException("Invalid token type");
            }
            JwtParser parser = getParser(secretKey);
            parser.parseSignedClaims(jwtToken);
        } catch (ExpiredJwtException eje) {
            logger.debug("Token is expired");
            return TokenState.EXPIRED;
        } catch (JwtException je) {
            logger.debug("Token is invalid");
            return TokenState.INVALID;
        } catch (Exception e) {
            throw new RuntimeException(STR."Unhandled exception: \{e.getMessage()}");
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
			case ACCESS_TOKEN -> secretKey = SECRET_ACCESS_KEY;
			case REFRESH_TOKEN -> secretKey = SECRET_REFRESH_KEY;
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
