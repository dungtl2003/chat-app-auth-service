package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.AllArgsConstructor;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.exception.InvalidUserException;
import org.service.auth.chatappauthservice.exception.UserNotFoundException;
import org.service.auth.chatappauthservice.service.UserService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@AllArgsConstructor
@RequestMapping("/api/v1/auth")
public class AuthController {

	private static final Logger logger = LogManager.getLogger(AuthController.class);

	private final UserService userService;

	@GetMapping("/bcrypt")
    public void createRandomUser(@RequestBody JsonNode request) {
        JsonNode payload = request.get("payload");
        logger.debug(STR."payload: \{request}");
        userService.createRandomUser();
    }

	@GetMapping("/authenticate")
    public void authenticate(@RequestBody JsonNode request) throws UserNotFoundException, InvalidUserException {
        logger.info(STR."Received request: \{request}");
        JsonNode jsonUser = request.get("payload").get("user");
        String email = jsonUser.get("email").asText();
        String password = jsonUser.get("password").asText();

        UserDTO user = userService.getValidUser(email, password);
    }

}
