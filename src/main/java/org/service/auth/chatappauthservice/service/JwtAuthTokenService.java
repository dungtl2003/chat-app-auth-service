package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.springframework.stereotype.Service;

@AllArgsConstructor
@Service
public class JwtAuthTokenService implements AuthTokenService {

	@Override
	public String createToken(UserDTO user) {
		return null;
	}

}
