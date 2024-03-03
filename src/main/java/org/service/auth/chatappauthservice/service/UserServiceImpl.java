package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import org.service.auth.chatappauthservice.repository.UserRepository;

@AllArgsConstructor
public class UserServiceImpl implements UserService {

	private final UserRepository userRepository;

}
