package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.databind.JsonNode;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestBody;

public interface AuthService {

	ResponseEntity<String> authenticate(@RequestBody JsonNode request)
			throws UserNotFoundException, InvalidUserException;

}
