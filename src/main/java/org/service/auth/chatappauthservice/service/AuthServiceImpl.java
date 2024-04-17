package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.AllArgsConstructor;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.entity.enums.TokenType;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.RequestHeader;

import java.util.Map;

@AllArgsConstructor
@Service
public class AuthServiceImpl implements AuthService {

	private final UserService userService;

	private final AuthTokenService authTokenService;

	@Override
	public ResponseEntity<AuthenticationResponse> login(JsonNode request)
			throws UserNotFoundException, InvalidUserException {
		JsonNode jsonUser = request.get("payload").get("user");
		String email = jsonUser.get("email").asText();
		String password = jsonUser.get("password").asText();

		User user = userService.getValidUser(email, password);

		String accessToken = authTokenService.createAccessToken(user);
		String refreshToken = authTokenService.createRefreshToken(user);

		userService.updateUserRefreshTokens(user, refreshToken);

		AuthenticationResponse responseBody = AuthenticationResponse.builder().accessToken(accessToken).build();

		HttpHeaders headers = new HttpHeaders();
		headers.setContentType(MediaType.APPLICATION_JSON);

		ResponseCookie cookie = ResponseCookie.from("jwt", refreshToken)
			// .secure(true) //TODO: add this when you have https
			.maxAge(7 * 24 * 60 * 60) // same as the refresh token
			.httpOnly(true)
			.path("/refresh")
			.build();

		headers.add(HttpHeaders.SET_COOKIE, cookie.toString());

		return new ResponseEntity<>(responseBody, headers, HttpStatus.OK);
	}

	@Override
	public ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers) {
		if (!headers.containsKey("authorization")) {
			AuthorizationResponse response = AuthorizationResponse.builder().message("Missing credential").build();
			return new ResponseEntity<>(response, HttpStatus.UNAUTHORIZED);
		}

		if (!headers.get("authorization").startsWith("Bearer ")) {
			AuthorizationResponse response = AuthorizationResponse.builder().message("Invalid format").build();
			return new ResponseEntity<>(response, HttpStatus.UNAUTHORIZED);
		}

		String token = headers.get("authorization").substring(7).strip();
		if (!authTokenService.isTokenValid(token, TokenType.ACCESS_TOKEN)) {
			AuthorizationResponse response = AuthorizationResponse.builder().message("Invalid token").build();
			return new ResponseEntity<>(response, HttpStatus.UNAUTHORIZED);
		}

		AuthorizationResponse response = AuthorizationResponse.builder().message("Authorized").build();
		return new ResponseEntity<>(response, HttpStatus.UNAUTHORIZED);
	}

}
