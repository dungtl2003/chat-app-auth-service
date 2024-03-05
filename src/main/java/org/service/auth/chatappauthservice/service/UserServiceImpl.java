package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.exception.InvalidUserException;
import org.service.auth.chatappauthservice.exception.UserNotFoundException;
import org.service.auth.chatappauthservice.repository.UserRepository;
import org.service.auth.chatappauthservice.utils.UserDTOMapper;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@AllArgsConstructor
@Service
public class UserServiceImpl implements UserService {

	private final UserRepository userRepository;

	private final UserDTOMapper userDTOMapper;

	private final PasswordEncoder passwordEncoder;

	// TODO: for testing purpose, must be removed after
	@Override
	public void createRandomUser() {
		String password = "password";
		String encoded = passwordEncoder.encode(password);
	}

	@Override
	public UserDTO getValidUser(String email, String password) throws UserNotFoundException, InvalidUserException {
		UserDTO user = getUserByEmail(email);

		if (!isValid(email, password, user)) {
			throw new InvalidUserException("Invalid email and password");
		}

		return user;
	}

	private UserDTO getUserByEmail(String email) throws UserNotFoundException {
        return userRepository
                .findByEmail(email)
                .map(userDTOMapper)
                .orElseThrow(() -> new UserNotFoundException(STR."User with email \{email} does not exist"));
    }

	private boolean isValid(String email, String password, UserDTO user) {
		return email.equals(user.email()) && passwordEncoder.matches(password, user.password());
	}

}
