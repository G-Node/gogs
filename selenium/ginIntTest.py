# -*- coding: utf-8 -*-
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support.ui import Select
from selenium.common.exceptions import NoSuchElementException
from selenium.common.exceptions import NoAlertPresentException
import unittest, time, re
import os

class GINLanding(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        ginurl = os.environ["GINURL"]
        cls.driver = webdriver.Chrome("bin/chromedriver")
        cls.driver.implicitly_wait(3)
        cls.base_url = ginurl
        cls.verificationErrors = []
        cls.accept_next_alert = True
    
    @classmethod
    def tearDownClass(cls):
        #cls.driver.quit()
        return
    
    def setUp(self):
        return
    def test00_install(self):
        driver = self.driver
        driver.get(self.base_url + "/install")
        driver.find_element_by_css_selector("div.ui.selection.database.type.dropdown").click()
        # some wait is needed
        time.sleep(1)
        driver.find_element_by_xpath("//div[4]").click()
        driver.find_element_by_id("db_path").clear()
        driver.find_element_by_id("db_path").send_keys("/data/gogs.db")
        driver.find_element_by_xpath("//div[@id='sqlite_settings']/div").click()
        driver.find_element_by_id("app_name").click()
        driver.find_element_by_id("app_name").clear()
        driver.find_element_by_id("app_name").send_keys("GINTEST")
        driver.find_element_by_id("app_url").click()
        driver.find_element_by_id("app_url").clear()
        driver.find_element_by_id("app_url").send_keys(self.base_url + "")
        driver.find_element_by_css_selector("button.ui.primary.button").click()
        self.assertEqual("Sign In", driver.find_element_by_css_selector("h3.ui.top.attached.header").text)

    def test01_g_i_n_landing(self):
        driver = self.driver
        driver.get(self.base_url + "")
        driver.find_element_by_link_text("Home").click()
        self.assertEqual("Modern Research Data Management for Neuroscience", driver.find_element_by_css_selector("h2").text)
        self.assertEqual("...inspired by github, flavoured for science", driver.find_element_by_css_selector("div.ginsubtitle").text)
        self.assertEqual("FAQ", driver.find_element_by_link_text("FAQ").text)
        self.assertEqual("Register", driver.find_element_by_link_text("Register").text)
        self.assertEqual("Sign In", driver.find_element_by_link_text("Sign In").text)
        self.assertEqual("GINTEST", driver.title)
    
    def test02_register(self):
        driver = self.driver
        driver.get(self.base_url + "")
        driver.find_element_by_link_text("Register").click()
        self.assertEqual("Sign Up", driver.find_element_by_css_selector("h3.ui.top.attached.header").text)
        self.assertEqual("Please note!\nFor Registration we require only username, password and email. Please use an institutional email to register. Otherwise you will only be able to use a subset of gins functionality and your maximum repository size will be dramatically reduced", driver.find_element_by_css_selector("div.ui.piled.yellow.segment").text)
        self.assertEqual("", driver.find_element_by_css_selector("button.ui.green.button").get_attribute("value"))
        driver.find_element_by_id("user_name").click()
        driver.find_element_by_id("user_name").clear()
        driver.find_element_by_id("user_name").send_keys("testuser")
        driver.find_element_by_id("email").clear()
        driver.find_element_by_id("email").send_keys("test@test.test")
        driver.find_element_by_id("password").clear()
        driver.find_element_by_id("password").send_keys("test")
        driver.find_element_by_id("retype").clear()
        driver.find_element_by_id("retype").send_keys("test")
        driver.find_element_by_id("full_name").clear()
        driver.find_element_by_id("full_name").send_keys("test")
        driver.find_element_by_id("affiliation").clear()
        driver.find_element_by_id("affiliation").send_keys("tester")
        driver.find_element_by_css_selector("button.ui.green.button").click()
        self.assertEqual("Sign In", driver.find_element_by_css_selector("h3.ui.top.attached.header").text)


    def test03_login(self):
        driver = self.driver
        self.login()
        self.assertEqual("test - Dashboard - GINTEST", driver.title)

    def test04_createrepo(self):
        driver = self.driver
        driver.get(self.base_url + "")
        #driver.find_element_by_css_selector("div.full.height").click()
        driver.find_element_by_css_selector("i.octicon.octicon-triangle-down").click()
        #driver.find_element_by_css_selector("div.ui.dropdown.head.link.jump.item.poping.up.visible").click()
        self.assertEqual("New Repository", driver.find_element_by_link_text("New Repository").text)
        driver.find_element_by_link_text("New Repository").click()
        self.assertEqual("New Repository - GINTEST", driver.title)
        driver.find_element_by_id("repo_name").click()
        driver.find_element_by_id("repo_name").clear()
        driver.find_element_by_id("repo_name").send_keys("testrepo1")
        driver.find_element_by_id("description").clear()
        driver.find_element_by_id("description").send_keys("this is the first test repository")
        driver.find_element_by_css_selector("button.ui.green.button").click()
        self.assertEqual("testuser/testrepo1: this is the first test repository - GINTEST", driver.title)
        self.assertEqual("LICENSE", driver.find_element_by_link_text("LICENSE").text)
        self.assertEqual("README.md", driver.find_element_by_link_text("README.md").text)

    def is_element_present(self, how, what):
        try: self.driver.find_element(by=how, value=what)
        except NoSuchElementException as e: return False
        return True
    
    def is_alert_present(self):
        try: self.driver.switch_to_alert()
        except NoAlertPresentException as e: return False
        return True
    
    def close_alert_and_get_its_text(self):
        try:
            alert = self.driver.switch_to_alert()
            alert_text = alert.text
            if self.accept_next_alert:
                alert.accept()
            else:
                alert.dismiss()
            return alert_text
        finally: self.accept_next_alert = True
    
    def tearDown(self):
        return


    def logout(self):
        """
        Helper to log out of current session
        """
        driver = self.driver
        driver.get(self.base_url + "/user/logout")
    
    def login(self):
        """
        Helper to login Testuser
        """
        driver = self.driver
        driver.get(self.base_url + "/user/login")
        self.assertEqual("Sign In", driver.find_element_by_css_selector("h3.ui.top.attached.header").text)
        self.assertEqual("Sign In", driver.find_element_by_css_selector("button.ui.green.button").text)
        driver.find_element_by_id("user_name").click()
        driver.find_element_by_id("user_name").clear()
        driver.find_element_by_id("user_name").send_keys("testuser")
        driver.find_element_by_id("password").clear()
        driver.find_element_by_id("password").send_keys("test")
        driver.find_element_by_css_selector("button.ui.green.button").click()

if __name__ == "__main__":
    unittest.main()
