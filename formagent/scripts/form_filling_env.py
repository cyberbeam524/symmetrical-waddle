import gym
from gym import spaces
import numpy as np

class FormFillingEnv(gym.Env):
    def __init__(self):
        super().__init__()
        self.state = None
        self.fields = ["name", "email", "resume"]
        self.actions = ["fill", "upload", "submit"]

        self.action_space = spaces.Discrete(len(self.actions))
        self.observation_space = spaces.Discrete(len(self.fields))

    def reset(self):
        self.state = 0
        return self.state

    def step(self, action):
        if action == 0:  # Fill
            reward = 1 if self.state < len(self.fields) else -1
        elif action == 1:  # Upload
            reward = 1 if "resume" in self.fields else -1
        elif action == 2:  # Submit
            reward = 1 if self.state == len(self.fields) else -1
        else:
            reward = -1

        self.state += 1
        done = self.state >= len(self.fields)
        return self.state, reward, done, {}

env = FormFillingEnv()
