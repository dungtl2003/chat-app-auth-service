package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.databind.JsonNode;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;

import java.util.Map;

public interface AuthService {

	ResponseEntity<AuthenticationResponse> login(@RequestBody JsonNode request)
			throws UserNotFoundException, InvalidUserException;

	ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers);

}
