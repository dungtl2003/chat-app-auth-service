package org.service.auth.chatappauthservice.service;

import org.service.auth.chatappauthservice.DTO.UserDTO;

public interface AuthTokenService {

	public String createToken(UserDTO user, long expiration);

}
