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
import org.springframework.beans.factory.annotation.Value;
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

	private static final long DEFAULT_ACCESS_EXPIRATION = 2 * 60 * 60 * 1000; // 2 hours

	private static final long REFRESH_EXPIRATION = 7 * 24 * 60 * 60 * 1000; // 7 days

	@Value("${application.jwt.access_token_secret}")
	private String secretAccessKey;

	@Value("${application.jwt.refresh_token_secret}")
	private String secretRefreshKey;

	public JwtAuthTokenService() {
	}

	@Override
	public String createAccessToken(UserDTO user) {
		return createToken(user, secretAccessKey, DEFAULT_ACCESS_EXPIRATION);
	}

	@Override
	public String createRefreshToken(UserDTO user) {
		return createToken(user, secretRefreshKey, REFRESH_EXPIRATION);
	}

	@Override
	public String createToken(UserDTO user, long expiration) {
		return createToken(user, secretRefreshKey, expiration);
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
			.add("role", user.role())
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
                case ACCESS_TOKEN -> secretKey = secretAccessKey;
                case REFRESH_TOKEN -> secretKey = secretRefreshKey;
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
			case ACCESS_TOKEN -> secretKey = secretAccessKey;
			case REFRESH_TOKEN -> secretKey = secretRefreshKey;
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
