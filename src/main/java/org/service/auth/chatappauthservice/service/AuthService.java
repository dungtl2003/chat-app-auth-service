package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import org.service.auth.chatappauthservice.exception.authorize.InvalidAuthorizationHeaderException;
import org.service.auth.chatappauthservice.exception.authorize.MissingAccessTokenException;
import org.service.auth.chatappauthservice.exception.refresh.MissingRefreshTokenException;
import org.service.auth.chatappauthservice.exception.token.InvalidTokenException;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;

import java.util.Map;

public interface AuthService {

	ResponseEntity<AuthenticationResponse> login(@RequestBody JsonNode request)
			throws UserNotFoundException, InvalidUserException;

	ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers)
			throws MissingAccessTokenException, InvalidAuthorizationHeaderException, InvalidTokenException;

	ResponseEntity<RefreshResponse> refresh(@RequestHeader Map<String, String> cookies)
			throws MissingRefreshTokenException, InvalidTokenException, JsonProcessingException;

}
