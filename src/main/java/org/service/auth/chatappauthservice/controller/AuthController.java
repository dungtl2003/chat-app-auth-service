package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.AllArgsConstructor;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.service.auth.chatappauthservice.service.AuthService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@AllArgsConstructor
@RequestMapping("/api/v1/auth")
public class AuthController {

	private static final Logger logger = LogManager.getLogger(AuthController.class);

	private final AuthService authService;

	@PostMapping("/login")
    public ResponseEntity<String> authenticate(@RequestBody JsonNode request) throws UserNotFoundException, InvalidUserException {
        logger.info(STR."Received request: \{request}");
        return authService.login(request);
    }

	// TODO: do later
	public ResponseEntity<String> authorize(@RequestBody JsonNode request) {
        logger.info(STR."Received request: \{request}");
        return authService.authorize(request);
    }

}
