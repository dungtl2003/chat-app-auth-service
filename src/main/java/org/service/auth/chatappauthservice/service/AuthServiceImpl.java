package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.entity.enums.TokenState;
import org.service.auth.chatappauthservice.entity.enums.TokenType;
import org.service.auth.chatappauthservice.exception.authorize.InvalidAuthorizationHeaderException;
import org.service.auth.chatappauthservice.exception.authorize.MissingAccessTokenException;
import org.service.auth.chatappauthservice.exception.refresh.MissingRefreshTokenException;
import org.service.auth.chatappauthservice.exception.token.ExpiredTokenException;
import org.service.auth.chatappauthservice.exception.token.InvalidTokenException;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.service.auth.chatappauthservice.utils.UserDTOMapper;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.RequestHeader;

import java.util.Arrays;
import java.util.Map;
import java.util.Objects;

@Service
public class AuthServiceImpl implements AuthService {

	private final UserService userService;

	private final AuthTokenService authTokenService;

	@Value("${info.app.version}")
	private int apiVersion;

	public AuthServiceImpl(UserService userService, AuthTokenService authTokenService) {
		this.userService = userService;
		this.authTokenService = authTokenService;
	}

	@Override
	public ResponseEntity<AuthenticationResponse> login(JsonNode request)
			throws UserNotFoundException, InvalidUserException {
		JsonNode jsonUser = request.get("payload").get("user");
		String email = jsonUser.get("email").asText();
		String password = jsonUser.get("password").asText();

		User user = userService.getValidUser(email, password);
		UserDTOMapper userDTOMapper = new UserDTOMapper();

		String accessToken = authTokenService.createAccessToken(userDTOMapper.apply(user));
		String refreshToken = authTokenService.createRefreshToken(userDTOMapper.apply(user));

		userService.addRefreshToken(user.getUserId(), refreshToken);

		AuthenticationResponse responseBody = AuthenticationResponse.builder()
			.podIP(Objects.equals(System.getenv("ENVIRONMENT"), "development") ? System.getenv("MY_POD_ID") : null)
			.accessToken(accessToken)
			.build();

		HttpHeaders headers = new HttpHeaders();
		headers.setContentType(MediaType.APPLICATION_JSON);
		headers.add(HttpHeaders.SET_COOKIE, getRefreshTokenCookie(refreshToken).toString());

		return new ResponseEntity<>(responseBody, headers, HttpStatus.OK);
	}

	@Override
	public ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers)
			throws InvalidAuthorizationHeaderException, MissingAccessTokenException, InvalidTokenException {
		if (!headers.containsKey("authorization")) {
			throw new MissingAccessTokenException("Missing credential");
		}

		if (!headers.get("authorization").startsWith("Bearer ")) {
			throw new InvalidAuthorizationHeaderException("Invalid format");
		}

		String token = headers.get("authorization").substring(7).strip();
		if (!authTokenService.checkTokenState(token, TokenType.ACCESS_TOKEN).equals(TokenState.VALID)) {
			throw new InvalidTokenException("Invalid token");
		}

		AuthorizationResponse response = AuthorizationResponse.builder()
			.podIP(Objects.equals(System.getenv("ENVIRONMENT"), "development") ? System.getenv("MY_POD_ID") : null)
			.message("Authorized")
			.build();
		return new ResponseEntity<>(response, HttpStatus.OK);
	}

	// TODO: hacker keeps sending this "valid" token and the user need to login again each
	// time he/she need to refresh tokens
	// SOLUTION 1: if the token is expired, even if it seems like being hacked, the server
	// just ignores
	@Override
	public ResponseEntity<RefreshResponse> refresh(Map<String, String> cookies)
			throws MissingRefreshTokenException, InvalidTokenException, JsonProcessingException {
		if (!cookies.containsKey("refresh_token")) {
			throw new MissingRefreshTokenException("Missing refresh token");
		}

		String refreshToken = cookies.get("refresh_token");
		TokenState state = authTokenService.checkTokenState(refreshToken, TokenType.REFRESH_TOKEN);

		// header, payload and signature does not match
		if (state.equals(TokenState.INVALID)) {
			throw new InvalidTokenException("Invalid refresh token");
		}

		// just in case the token is expired
		Map<String, String> body = authTokenService.parseBody(refreshToken);
		long userId = Long.parseLong(body.get("sub"));

		// get user's all current active refresh tokens
		String[] refreshTokens = userService.getUserRefreshTokens(userId);
		int oldLength = refreshTokens.length;

		refreshTokens = Arrays.stream(refreshTokens)
			.filter(token -> !token.equals(refreshToken))
			.toList()
			.toArray(new String[0]);
		boolean found = refreshTokens.length < oldLength;

		// if not found but expired, follow SOLUTION 1
		// else, remove all active refresh tokens so this user must log in again
		// in all devices to prevent hacker to use stolen refresh token
		userService.updateUserRefreshTokens(userId,
				found || state.equals(TokenState.EXPIRED) ? refreshTokens : new String[0]);

		// make this user to log in again in this device because the token is
		// expired
		if (state.equals(TokenState.EXPIRED)) {
			throw new ExpiredTokenException("Token is expired");
		}

		if (!found) {
			throw new ExpiredTokenException("Invalid refresh token");
		}

		// this token down here must be valid
		String email = authTokenService.extractClaim(refreshToken, claims -> claims.get("email", String.class),
				TokenType.REFRESH_TOKEN);
		String username = authTokenService.extractClaim(refreshToken, claims -> claims.get("username", String.class),
				TokenType.REFRESH_TOKEN);
		String role = authTokenService.extractClaim(refreshToken, claims -> claims.get("role", String.class),
				TokenType.REFRESH_TOKEN);
		UserDTO user = new UserDTO(userId, email, username, role);

		String newAccessToken = authTokenService.createAccessToken(user);
		String newRefreshToken = authTokenService.createRefreshToken(user);

		userService.addRefreshToken(user.userId(), newRefreshToken);

		RefreshResponse response = RefreshResponse.builder()
			.podIP(Objects.equals(System.getenv("ENVIRONMENT"), "development") ? System.getenv("MY_POD_ID") : null)
			.accessToken(newAccessToken)
			.build();

		HttpHeaders headers = new HttpHeaders();
		headers.setContentType(MediaType.APPLICATION_JSON);
		headers.add(HttpHeaders.SET_COOKIE, getRefreshTokenCookie(newRefreshToken).toString());

		return new ResponseEntity<>(response, headers, HttpStatus.OK);
	}

	private ResponseCookie getRefreshTokenCookie(String refreshToken) {
        return ResponseCookie.from("refresh_token", refreshToken)
                // .secure(true) //TODO: add this when you have https
                .maxAge(7 * 24 * 60 * 60) // same as the refresh token
                .httpOnly(true)
                .path(STR."/api/v\{apiVersion}/auth/refresh")
                .build();
    }

}
