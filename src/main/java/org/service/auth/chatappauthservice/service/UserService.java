package org.service.auth.chatappauthservice.service;

import org.service.auth.chatappauthservice.DTO.UserDTO;

public interface UserService {

	public void createRandomUser();

	public UserDTO getValidUser(String email, String password);

}
