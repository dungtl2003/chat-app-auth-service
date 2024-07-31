package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import lombok.AllArgsConstructor;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.exception.authorize.InvalidAuthorizationHeaderException;
import org.service.auth.chatappauthservice.exception.authorize.MissingAccessTokenException;
import org.service.auth.chatappauthservice.exception.refresh.MissingRefreshTokenException;
import org.service.auth.chatappauthservice.exception.refresh.ReusedRefreshTokenException;
import org.service.auth.chatappauthservice.exception.token.ExpiredTokenException;
import org.service.auth.chatappauthservice.exception.token.InvalidTokenException;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.LogoutResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.service.auth.chatappauthservice.service.AuthService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

@RestController
@AllArgsConstructor
@RequestMapping("/api/v1/auth")
public class AuthController implements AuthApi {

	private final AuthService authService;

	@Override
	public ResponseEntity<AuthenticationResponse> authenticate(@RequestBody JsonNode request)
			throws UserNotFoundException, InvalidUserException {
		return authService.login(request);
	}

	@Override
	public ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers)
			throws MissingAccessTokenException, InvalidAuthorizationHeaderException, InvalidTokenException,
			UserNotFoundException {
		return authService.authorize(headers);
	}

	@Override
	public ResponseEntity<RefreshResponse> refresh(HttpServletRequest request)
			throws MissingRefreshTokenException, InvalidTokenException, JsonProcessingException,
			ReusedRefreshTokenException, UserNotFoundException, ExpiredTokenException {
		Cookie[] rawCookies = request.getCookies();
		Map<String, String> cookies = new HashMap<>();

		if (rawCookies != null) {
			Arrays.stream(rawCookies)
				.forEach(rawCookie -> cookies.merge(rawCookie.getName(), rawCookie.getValue(),
						(before, after) -> after));
		}

		return authService.refresh(cookies);
	}

	@Override
	public ResponseEntity<LogoutResponse> logout(HttpServletRequest request)
			throws InvalidTokenException, ExpiredTokenException, MissingRefreshTokenException, UserNotFoundException {
		Cookie[] rawCookies = request.getCookies();
		Map<String, String> cookies = new HashMap<>();

		if (rawCookies != null) {
			Arrays.stream(rawCookies)
				.forEach(rawCookie -> cookies.merge(rawCookie.getName(), rawCookie.getValue(),
						(before, after) -> after));
		}

		return authService.logout(cookies);
	}

}
