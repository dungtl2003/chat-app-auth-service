package org.service.auth.chatappauthservice.utils;

import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.entity.User;

import java.util.function.Function;

public class UserDTOMapper implements Function<User, UserDTO> {

	@Override
	public UserDTO apply(User user) {
		if (user == null) {
			return null;
		}
		return new UserDTO(user.getUserId(), user.getEmail(), user.getUsername(), user.getRole().getName());
	}

}